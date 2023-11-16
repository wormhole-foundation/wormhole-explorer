package relays

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Repository struct {
	db          *mongo.Database
	logger      *zap.Logger
	collections struct {
		relays *mongo.Collection
	}
}

func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger: logger.With(zap.String("module", "VaaRepository")),
		collections: struct {
			relays *mongo.Collection
		}{
			relays: db.Collection("relays"),
		},
	}
}

func (r *Repository) FindOne(ctx context.Context, q *RelaysQuery) (*RelayDoc, error) {
	var response RelayDoc
	err := r.collections.relays.FindOne(ctx, q.toBSON()).Decode(&response)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrNotFound
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get relays",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	return &response, nil
}

type RelaysQuery struct {
	chainId  vaa.ChainID
	emitter  string
	sequence string
}

type RelayDoc struct {
	ID   string `bson:"_id"`
	Data struct {
		Status      string     `bson:"status"`
		ReceivedAt  time.Time  `bson:"receivedAt"`
		CompletedAt *time.Time `bson:"completedAt"`
		ToTxHash    *string    `bson:"toTxHash"`
		Metadata    *struct {
			Attempts                 int   `bson:"attempts"`
			ExecutionStartTime       int64 `bson:"executionStartTime"`
			EmitterChain             int   `bson:"emitterChain"`
			DidMatchDeliveryProvider bool  `bson:"didMatchDeliveryProvider"`
			DidParse                 bool  `bson:"didParse"`
			Instructions             struct {
				EncodedExecutionInfo   string `bson:"encodedExecutionInfo"`
				RefundAddress          string `bson:"refundAddress"`
				SourceDeliveryProvider string `bson:"sourceDeliveryProvider"`
				SenderAddress          string `bson:"senderAddress"`
				VaaKeys                []any  `bson:"vaaKeys"`
				ExtraReceiverValue     struct {
					Hex         string `bson:"_hex"`
					IsBigNumber bool   `bson:"_isBigNumber"`
				} `bson:"extraReceiverValue"`
				TargetAddress          string `bson:"targetAddress"`
				RequestedReceiverValue struct {
					Hex         string `bson:"_hex"`
					IsBigNumber bool   `bson:"_isBigNumber"`
				} `bson:"requestedReceiverValue"`
				RefundChainID          int    `bson:"refundChainId"`
				RefundDeliveryProvider string `bson:"refundDeliveryProvider"`
				TargetChainID          int    `bson:"targetChainId"`
			} `bson:"instructions"`
			DeliveryRecord struct {
				MaxRefund                   string   `bson:"maxRefund"`
				Budget                      string   `bson:"budget"`
				TargetChainAssetPriceUSD    float64  `bson:"targetChainAssetPriceUSD"`
				WalletNonce                 int      `bson:"walletNonce"`
				TransactionHashes           []string `bson:"transactionHashes"`
				HasAdditionalVaas           bool     `bson:"hasAdditionalVaas"`
				AdditionalVaasDidFetch      bool     `bson:"additionalVaasDidFetch"`
				WalletAcquisitionEndTime    int64    `bson:"walletAcquisitionEndTime"`
				WalletAcquisitionDidSucceed bool     `bson:"walletAcquisitionDidSucceed"`
				WalletBalanceAfter          string   `bson:"walletBalanceAfter"`
				ResultLog                   struct {
					TransactionHash   string `bson:"transactionHash"`
					VaaHash           string `bson:"vaaHash"`
					RefundStatus      string `bson:"refundStatus"`
					RevertString      string `bson:"revertString"`
					Status            string `bson:"status"`
					GasUsed           string `bson:"gasUsed"`
					SourceChain       string `bson:"sourceChain"`
					SourceVaaSequence string `bson:"sourceVaaSequence"`
				} `bson:"resultLog"`
				ResultString                  string  `bson:"resultString"`
				AdditionalVaaKeysPrintable    string  `bson:"additionalVaaKeysPrintable"`
				BudgetUsd                     float64 `bson:"budgetUsd"`
				WalletAcquisitionStartTime    int64   `bson:"walletAcquisitionStartTime"`
				GasUnitsEstimate              int     `bson:"gasUnitsEstimate"`
				EstimatedTransactionFeeEther  string  `bson:"estimatedTransactionFeeEther"`
				TargetChainDecimals           int     `bson:"targetChainDecimals"`
				DeliveryInstructionsPrintable struct {
					Payload                string `bson:"payload"`
					EncodedExecutionInfo   string `bson:"encodedExecutionInfo"`
					RefundDeliveryProvider string `bson:"refundDeliveryProvider"`
					SourceDeliveryProvider string `bson:"sourceDeliveryProvider"`
					SenderAddress          string `bson:"senderAddress"`
					TargetAddress          string `bson:"targetAddress"`
					RequestedReceiverValue string `bson:"requestedReceiverValue"`
					ExtraReceiverValue     string `bson:"extraReceiverValue"`
					RefundChainID          string `bson:"refundChainId"`
					RefundAddress          string `bson:"refundAddress"`
					VaaKeys                []any  `bson:"vaaKeys"`
					TargetChainID          string `bson:"targetChainId"`
				} `bson:"deliveryInstructionsPrintable"`
				WalletAddress              string  `bson:"walletAddress"`
				GasUsed                    int     `bson:"gasUsed"`
				GasPrice                   string  `bson:"gasPrice"`
				ReceiverValue              string  `bson:"receiverValue"`
				MaxRefundUsd               float64 `bson:"maxRefundUsd"`
				GasPriceEstimate           string  `bson:"gasPriceEstimate"`
				TransactionDidSubmit       bool    `bson:"transactionDidSubmit"`
				EstimatedTransactionFee    string  `bson:"estimatedTransactionFee"`
				TransactionSubmitTimeStart int64   `bson:"transactionSubmitTimeStart"`
				TransactionSubmitTimeEnd   int64   `bson:"transactionSubmitTimeEnd"`
				ResultLogDidParse          bool    `bson:"resultLogDidParse"`
				ChainID                    int     `bson:"chainId"`
				ReceiverValueUsd           float64 `bson:"receiverValueUsd"`
				WalletBalanceBefore        string  `bson:"walletBalanceBefore"`
			} `bson:"deliveryRecord"`
			RawVaaHex            string `bson:"rawVaaHex"`
			PayloadType          int    `bson:"payloadType"`
			MaxAttempts          int    `bson:"maxAttempts"`
			DidError             bool   `bson:"didError"`
			ExecutionEndTime     int64  `bson:"executionEndTime"`
			EmitterAddress       string `bson:"emitterAddress"`
			DidSubmitTransaction bool   `bson:"didSubmitTransaction"`
			Sequence             string `bson:"sequence"`
		} `bson:"metadata"`
		Sequence       string     `bson:"sequence"`
		Vaa            string     `bson:"vaa"`
		FromTxHash     string     `bson:"fromTxHash"`
		MaxAttempts    int        `bson:"maxAttempts"`
		AddedTimes     int        `bson:"addedTimes"`
		ErrorMessage   any        `bson:"errorMessage"`
		EmitterChain   int        `bson:"emitterChain"`
		EmitterAddress string     `bson:"emitterAddress"`
		FailedAt       *time.Time `bson:"failedAt"`
	} `bson:"data"`
	Event  string `bson:"event"`
	Origin string `bson:"origin"`
}

func (q *RelaysQuery) toBSON() *bson.D {
	r := bson.D{}
	id := fmt.Sprintf("%d/%s/%s", q.chainId, q.emitter, q.sequence)
	r = append(r, bson.E{Key: "_id", Value: id})
	return &r
}

func (q *RelaysQuery) SetChain(chainId vaa.ChainID) *RelaysQuery {
	q.chainId = chainId
	return q
}

func (q *RelaysQuery) SetEmitter(emitter string) *RelaysQuery {
	q.emitter = emitter
	return q
}

func (q *RelaysQuery) SetSequence(sequence string) *RelaysQuery {
	q.sequence = sequence
	return q
}

func Query() *RelaysQuery {
	return &RelaysQuery{}
}
