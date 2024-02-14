package report

import (
	"context"
	"encoding/csv"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/prices"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type TransferReportJob struct {
	database       *mongo.Database
	pageSize       int64
	logger         *zap.Logger
	getPriceByTime GetPriceByTimeFn
	outputPath     string
	tokenProvider  *domain.TokenProvider
}

type transactionResult struct {
	ID                         string      `bson:"_id" json:"vaaId"`
	SourceChain                sdk.ChainID `bson:"sourceChain" json:"sourceChain"`
	EmitterAddress             string      `bson:"emitterAddress" json:"emitterAddress"`
	Sequence                   string      `bson:"sequence" json:"sequence"`
	VaaHash                    string      `bson:"vaaHash" json:"vaaHash"`
	SourceTxHash               string      `bson:"sourceTxHash" json:"sourceTxHash"`
	DestinationChain           sdk.ChainID `bson:"destinationChain" json:"destinationChain"`
	DestinationTxHash          string      `bson:"destinationTxHash" json:"destinationTxHash"`
	TokenChain                 sdk.ChainID `bson:"tokenChain" json:"tokenChain"`
	TokenAddress               string      `bson:"tokenAddress" json:"tokenAddress"`
	TokenAddressHexa           string      `bson:"tokenAddressHexa" json:"tokenAddressHexa"`
	Amount                     string      `bson:"amount" json:"amount"`
	SourceSenderAddress        string      `bson:"sourceWallet" json:"sourceSenderAddress"`
	DestinationAddress         string      `bson:"destinationWallet" json:"destinationAddress"`
	Fee                        string      `bson:"fee" json:"fee"`
	Timestamp                  time.Time   `bson:"timestamp" json:"timestamp"`
	AppIds                     []string    `bson:"appIds" json:"appIds"`
	PortalPayloadType          int         `bson:"portalPayloadType" json:"portalPayloadType"`
	SemanticDestinationAddress string      `bson:"semanticDestinationAddress" json:"semanticDestinationAddress"`
}

type GetPriceByTimeFn func(ctx context.Context, coingeckoID string, day time.Time) (decimal.Decimal, error)

// NewTransferReportJob creates a new transfer report job.
func NewTransferReportJob(database *mongo.Database, pageSize int64, getPriceByTime GetPriceByTimeFn, outputPath string, tokenProvider *domain.TokenProvider, logger *zap.Logger) *TransferReportJob {
	return &TransferReportJob{database: database, pageSize: pageSize, getPriceByTime: getPriceByTime, outputPath: outputPath, tokenProvider: tokenProvider, logger: logger}
}

// Run runs the transfer report job.
func (j *TransferReportJob) Run(ctx context.Context) error {

	file, err := os.Create(j.outputPath)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)

	_ = j.writeHeader(writer)

	defer file.Close()

	//start backfilling
	page := int64(0)
	for {
		j.logger.Info("Processing page", zap.Int64("page", page))

		trxs, err := j.findTransactionsByPage(ctx, page, j.pageSize)
		if err != nil {
			j.logger.Error("Failed to get transactions", zap.Error(err))
			break
		}

		if len(trxs) == 0 {
			j.logger.Info("Empty page", zap.Int64("page", page))
			break
		}
		for _, t := range trxs {

			log := j.logger.With(zap.String("id", t.ID))

			log.Debug("Processing transaction")

			if t.TokenAddressHexa == "" {
				j.writeRecord(t, t.Amount, nil, nil, writer)
				continue
			}

			tokenAddress, err := sdk.StringToAddress(t.TokenAddressHexa)
			if err != nil {
				log.Error("Failed to get transactions",
					zap.String("tokenAddressHexa", t.TokenAddressHexa),
					zap.Error(err))
				continue
			}

			m, ok := j.tokenProvider.GetTokenByAddress(sdk.ChainID(t.TokenChain), tokenAddress.String())
			if ok {
				tokenPrice, err := j.getPriceByTime(ctx, m.CoingeckoID, t.Timestamp)
				if err != nil {
					log.Error("Failed to get token price",
						zap.String("coingeckoId", m.CoingeckoID),
						zap.String("timestamp", t.Timestamp.UTC().Format(time.RFC3339)),
						zap.Error(err))
					continue
				}
				if t.Amount == "" {
					j.writeRecord(t, t.Amount, nil, nil, writer)
					continue
				}
				amount := new(big.Int)
				amount, ok := amount.SetString(t.Amount, 10)
				if !ok {
					log.Error("amount is not a number",
						zap.String("amount", t.Amount),
					)
					j.writeRecord(t, "", nil, nil, writer)
					continue
				}

				priceUSD := prices.CalculatePriceUSD(tokenPrice, amount, m.Decimals)

				j.writeRecord(t, t.Amount, m, &priceUSD, writer)
			} else {
				j.writeRecord(t, t.Amount, nil, nil, writer)
			}

		}
		writer.Flush()
		page++
	}
	return nil
}

func (j *TransferReportJob) writeRecord(trx transactionResult, fAmount string, m *domain.TokenMetadata, priceUSD *decimal.Decimal, file *csv.Writer) error {
	var notionalUSD, decimals, symbol, coingeckoID, tokenAddress string
	if m != nil {
		decimals = fmt.Sprintf("%d", m.Decimals)
		symbol = m.Symbol.String()
		coingeckoID = m.CoingeckoID
	}
	if priceUSD != nil {
		notionalUSD = priceUSD.Truncate(10).String()
	}

	tokenAddress = trx.TokenAddress
	if !regexp.MustCompile(`^[A-Za-z0-9]*$`).MatchString(tokenAddress) {
		tokenAddress, _ = domain.TranslateEmitterAddress(trx.TokenChain, trx.TokenAddressHexa)
	}

	var record []string
	record = append(record, trx.ID)
	record = append(record, trx.VaaHash)
	record = append(record, chainIDToCsv(trx.SourceChain))
	record = append(record, trx.EmitterAddress)
	record = append(record, trx.Sequence)
	record = append(record, trx.Timestamp.Format(time.RFC3339))
	record = append(record, trx.SourceTxHash)
	record = append(record, trx.SourceSenderAddress)
	record = append(record, chainIDToCsv(trx.DestinationChain))
	record = append(record, trx.DestinationAddress)
	record = append(record, trx.DestinationTxHash)
	record = append(record, portalPayloadTypeToCsv(trx.PortalPayloadType))
	record = append(record, appIdsToCsv(trx.AppIds))
	record = append(record, chainIDToCsv(trx.TokenChain))
	record = append(record, tokenAddress)
	record = append(record, fAmount)
	record = append(record, decimals)
	record = append(record, notionalUSD)
	record = append(record, trx.Fee)
	record = append(record, coingeckoID)
	record = append(record, symbol)
	return file.Write(record)
}

func (*TransferReportJob) writeHeader(writer *csv.Writer) error {
	var record []string
	record = append(record, "vaaId")
	record = append(record, "vaaHash")
	record = append(record, "sourceChain")
	record = append(record, "emitterAddress")
	record = append(record, "sequence")
	record = append(record, "timestamp")
	record = append(record, "sourceTxHash")
	record = append(record, "sourceSenderAddress")
	record = append(record, "destinationChain")
	record = append(record, "destinationAddress")
	record = append(record, "destinationTxHash")
	record = append(record, "portalPayloadType")
	record = append(record, "appIds")
	record = append(record, "tokenChain")
	record = append(record, "tokenAddress")
	record = append(record, "amount")
	record = append(record, "decimals")
	record = append(record, "notionalUSD")
	record = append(record, "fee")
	record = append(record, "coinGeckoId")
	record = append(record, "symbol")
	return writer.Write(record)
}

func (j *TransferReportJob) findTransactionsByPage(ctx context.Context, page, pageSize int64) ([]transactionResult, error) {

	vaas := j.database.Collection("vaas")

	skip := page * pageSize

	// Build the aggregation pipeline
	var pipeline mongo.Pipeline

	pipeline = append(pipeline, bson.D{
		{Key: "$sort", Value: bson.D{
			bson.E{Key: "timestamp", Value: -1},
			bson.E{Key: "_id", Value: -1},
		}},
	})

	pipeline = append(pipeline, bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "parsedVaa"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "parsedVaa"},
		}},
	})

	pipeline = append(pipeline, bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "globalTransactions"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "globalTransactions"},
		}},
	})

	// Skip initial results
	pipeline = append(pipeline, bson.D{
		{Key: "$skip", Value: skip},
	})

	// Limit size of results
	pipeline = append(pipeline, bson.D{
		{Key: "$limit", Value: pageSize},
	})

	// add nested fields
	pipeline = append(pipeline, bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "standardizedProperties", Value: bson.M{"$arrayElemAt": []interface{}{"$parsedVaa.rawStandardizedProperties", 0}}},
			{Key: "globalTransactions", Value: bson.M{"$arrayElemAt": []interface{}{"$globalTransactions", 0}}},
			{Key: "appIds", Value: bson.M{"$arrayElemAt": []interface{}{"$parsedVaa.appIds", 0}}},
			{Key: "parsedPayload", Value: bson.M{"$arrayElemAt": []interface{}{"$parsedVaa.parsedPayload", 0}}},
		}},
	})

	pipeline = append(pipeline, bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "appIds", Value: "$appIds"},
			{Key: "sourceChain", Value: "$emitterChain"},
			{Key: "emitterAddress", Value: "$emitterAddr"},
			{Key: "sequence", Value: "$sequence"},
			{Key: "vaaHash", Value: "$txHash"},
			{Key: "destinationChain", Value: "$standardizedProperties.toChain"},
			{Key: "tokenChain", Value: "$standardizedProperties.tokenChain"},
			{Key: "tokenAddress", Value: "$standardizedProperties.tokenAddress"},
			{Key: "amount", Value: "$standardizedProperties.amount"},
			{Key: "sourceWallet", Value: "$globalTransactions.originTx.from"},
			{Key: "sourceTxHash", Value: "$globalTransactions.originTx.nativeTxHash"},
			{Key: "destinationTxHash", Value: "$globalTransactions.destinationTx.txHash"},
			{Key: "destinationWallet", Value: "$standardizedProperties.toAddress"},
			{Key: "fee", Value: "$standardizedProperties.fee"},
			{Key: "timestamp", Value: "$timestamp"},
			{Key: "tokenAddressHexa", Value: "$parsedPayload.tokenAddress"},
			{Key: "portalPayloadType", Value: "$parsedPayload.payloadType"},
		}}})

	// Execute the aggregation pipeline
	cur, err := vaas.Aggregate(ctx, pipeline)
	if err != nil {
		j.logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return nil, err
	}

	// Read results from cursor
	var documents []transactionResult
	err = cur.All(ctx, &documents)
	if err != nil {
		j.logger.Error("failed to decode cursor", zap.Error(err))
		return nil, err
	}

	return documents, nil
}

func chainIDToCsv(chainID sdk.ChainID) string {

	if chainID.String() == sdk.ChainIDUnset.String() {
		return ""
	}
	return fmt.Sprintf("%d", int16(chainID))
}

func appIdsToCsv(appIds []string) string {
	if len(appIds) == 0 {
		return ""
	}
	return strings.Join(appIds, "|")
}

func portalPayloadTypeToCsv(portalPayloadType int) string {
	if portalPayloadType == 0 {
		return ""
	}
	return fmt.Sprintf("%d", portalPayloadType)
}
