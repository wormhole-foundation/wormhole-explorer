package report

import (
	"context"
	"encoding/csv"
	"fmt"
	"math/big"
	"os"
	"regexp"
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
	database      *mongo.Database
	pageSize      int64
	logger        *zap.Logger
	pricesCache   *prices.CoinPricesCache
	outputPath    string
	tokenProvider *domain.TokenProvider
}

type transactionResult struct {
	ID                string      `bson:"_id"`
	SourceChain       sdk.ChainID `bson:"sourceChain"`
	EmitterAddress    string      `bson:"emitterAddress"`
	Sequence          string      `bson:"sequence"`
	DestinationChain  sdk.ChainID `bson:"destinationChain"`
	TokenChain        sdk.ChainID `bson:"tokenChain"`
	TokenAddress      string      `bson:"tokenAddress"`
	TokenAddressHexa  string      `bson:"tokenAddressHexa"`
	Amount            string      `bson:"amount"`
	SourceWallet      string      `bson:"sourceWallet"`
	DestinationWallet string      `bson:"destinationWallet"`
	Fee               string      `bson:"fee"`
	Timestamp         time.Time   `bson:"timestamp"`
	AppIds            []string    `bson:"appIds"`
}

// NewTransferReportJob creates a new transfer report job.
func NewTransferReportJob(database *mongo.Database, pageSize int64, pricesCache *prices.CoinPricesCache, outputPath string, tokenProvider *domain.TokenProvider, logger *zap.Logger) *TransferReportJob {
	return &TransferReportJob{database: database, pageSize: pageSize, pricesCache: pricesCache, outputPath: outputPath, tokenProvider: tokenProvider, logger: logger}
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
			j.logger.Debug("Processing transaction", zap.String("id", t.ID))

			if t.TokenAddressHexa == "" {
				j.writeRecord(t, t.Amount, nil, nil, writer)
				continue
			}

			tokenAddress, err := sdk.StringToAddress(t.TokenAddressHexa)
			if err != nil {
				j.logger.Error("Failed to get transactions",
					zap.String("id", t.ID),
					zap.String("TokenAddressHexa", t.TokenAddressHexa),
					zap.Error(err))
				continue
			}

			m, ok := j.tokenProvider.GetTokenByAddress(sdk.ChainID(t.TokenChain), tokenAddress.String())
			if ok {
				tokenPrice, err := j.pricesCache.GetPriceByTime(m.CoingeckoID, t.Timestamp)
				if err != nil {
					continue
				}
				if t.Amount == "" {
					j.writeRecord(t, t.Amount, nil, nil, writer)
					continue
				}
				amount := new(big.Int)
				amount, ok := amount.SetString(t.Amount, 10)
				if !ok {
					j.logger.Error("amount is not a number",
						zap.String("id", t.ID),
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
	record = append(record, chainIDToCsv(trx.SourceChain))
	record = append(record, trx.EmitterAddress)
	record = append(record, trx.Sequence)
	record = append(record, trx.SourceWallet)
	record = append(record, chainIDToCsv(trx.DestinationChain))
	record = append(record, trx.DestinationWallet)
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
	record = append(record, "sourceChain")
	record = append(record, "emitterAddress")
	record = append(record, "sequence")
	record = append(record, "sourceWallet")
	record = append(record, "destinationChain")
	record = append(record, "destinationWallet")
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
			{Key: "destinationChain", Value: "$standardizedProperties.toChain"},
			{Key: "tokenChain", Value: "$standardizedProperties.tokenChain"},
			{Key: "tokenAddress", Value: "$standardizedProperties.tokenAddress"},
			{Key: "amount", Value: "$standardizedProperties.amount"},
			{Key: "sourceWallet", Value: "$globalTransactions.originTx.from"},
			{Key: "destinationWallet", Value: "$standardizedProperties.toAddress"},
			{Key: "fee", Value: "$standardizedProperties.fee"},
			{Key: "timestamp", Value: "$timestamp"},
			{Key: "tokenAddressHexa", Value: "$parsedPayload.tokenAddress"},
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
