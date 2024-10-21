package relays

import (
	"time"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type RelaysQuery struct {
	chainId  vaa.ChainID
	emitter  string
	sequence string
}

type RelayData struct {
	Status      string     `bson:"status" json:"status"`
	ReceivedAt  time.Time  `bson:"receivedAt" json:"receivedAt"`
	CompletedAt *time.Time `bson:"completedAt" json:"completedAt"`
	ToTxHash    *string    `bson:"toTxHash" json:"toTxHash"`
	Metadata    *struct {
		Attempts                 int   `bson:"attempts" json:"attempts"`
		ExecutionStartTime       int64 `bson:"executionStartTime" json:"executionStartTime"`
		EmitterChain             int   `bson:"emitterChain" json:"emitterChain"`
		DidMatchDeliveryProvider bool  `bson:"didMatchDeliveryProvider" json:"didMatchDeliveryProvider"`
		DidParse                 bool  `bson:"didParse" json:"didParse"`
		Instructions             struct {
			EncodedExecutionInfo   string `bson:"encodedExecutionInfo" json:"encodedExecutionInfo"`
			RefundAddress          string `bson:"refundAddress" json:"refundAddress"`
			SourceDeliveryProvider string `bson:"sourceDeliveryProvider" json:"sourceDeliveryProvider"`
			SenderAddress          string `bson:"senderAddress" json:"senderAddress"`
			VaaKeys                []any  `bson:"vaaKeys" json:"vaaKeys"`
			ExtraReceiverValue     struct {
				Hex         string `bson:"_hex" json:"_hex"`
				IsBigNumber bool   `bson:"_isBigNumber" json:"_isBigNumber"`
			} `bson:"extraReceiverValue"`
			TargetAddress          string `bson:"targetAddress" json:"targetAddress"`
			RequestedReceiverValue struct {
				Hex         string `bson:"_hex" json:"_hex"`
				IsBigNumber bool   `bson:"_isBigNumber" json:"_isBigNumber"`
			} `bson:"requestedReceiverValue"`
			RefundChainID          int    `bson:"refundChainId" json:"refundChainId"`
			RefundDeliveryProvider string `bson:"refundDeliveryProvider" json:"refundDeliveryProvider"`
			TargetChainID          int    `bson:"targetChainId" json:"targetChainId"`
		} `bson:"instructions" json:"instructions"`
		DeliveryRecord struct {
			MaxRefund                   string   `bson:"maxRefund" json:"maxRefund"`
			Budget                      string   `bson:"budget" json:"budget"`
			TargetChainAssetPriceUSD    float64  `bson:"targetChainAssetPriceUSD" json:"targetChainAssetPriceUSD"`
			WalletNonce                 int      `bson:"walletNonce" json:"walletNonce"`
			TransactionHashes           []string `bson:"transactionHashes" json:"transactionHashes"`
			HasAdditionalVaas           bool     `bson:"hasAdditionalVaas" json:"hasAdditionalVaas"`
			AdditionalVaasDidFetch      bool     `bson:"additionalVaasDidFetch" json:"additionalVaasDidFetch"`
			WalletAcquisitionEndTime    int64    `bson:"walletAcquisitionEndTime" json:"walletAcquisitionEndTime"`
			WalletAcquisitionDidSucceed bool     `bson:"walletAcquisitionDidSucceed" json:"walletAcquisitionDidSucceed"`
			WalletBalanceAfter          string   `bson:"walletBalanceAfter" json:"walletBalanceAfter"`
			ResultLog                   struct {
				TransactionHash   string `bson:"transactionHash" json:"transactionHash"`
				VaaHash           string `bson:"vaaHash" json:"vaaHash"`
				RefundStatus      string `bson:"refundStatus" json:"refundStatus"`
				RevertString      string `bson:"revertString" json:"revertString"`
				Status            string `bson:"status" json:"status"`
				GasUsed           string `bson:"gasUsed" json:"gasUsed"`
				SourceChain       string `bson:"sourceChain" json:"sourceChain"`
				SourceVaaSequence string `bson:"sourceVaaSequence" json:"sourceVaaSequence"`
			} `bson:"resultLog" json:"resultLog"`
			ResultString                  string  `bson:"resultString" json:"resultString"`
			AdditionalVaaKeysPrintable    string  `bson:"additionalVaaKeysPrintable" json:"additionalVaaKeysPrintable"`
			BudgetUsd                     float64 `bson:"budgetUsd" json:"budgetUsd"`
			WalletAcquisitionStartTime    int64   `bson:"walletAcquisitionStartTime" json:"walletAcquisitionStartTime"`
			GasUnitsEstimate              int     `bson:"gasUnitsEstimate" json:"gasUnitsEstimate"`
			EstimatedTransactionFeeEther  string  `bson:"estimatedTransactionFeeEther" json:"estimatedTransactionFeeEther"`
			TargetChainDecimals           int     `bson:"targetChainDecimals" json:"targetChainDecimals"`
			DeliveryInstructionsPrintable struct {
				Payload                string `bson:"payload" json:"payload"`
				EncodedExecutionInfo   string `bson:"encodedExecutionInfo" json:"encodedExecutionInfo"`
				RefundDeliveryProvider string `bson:"refundDeliveryProvider" json:"refundDeliveryProvider"`
				SourceDeliveryProvider string `bson:"sourceDeliveryProvider" json:"sourceDeliveryProvider"`
				SenderAddress          string `bson:"senderAddress" json:"senderAddress"`
				TargetAddress          string `bson:"targetAddress" json:"targetAddress"`
				RequestedReceiverValue string `bson:"requestedReceiverValue" json:"requestedReceiverValue"`
				ExtraReceiverValue     string `bson:"extraReceiverValue" json:"extraReceiverValue"`
				RefundChainID          string `bson:"refundChainId" json:"refundChainId"`
				RefundAddress          string `bson:"refundAddress" json:"refundAddress"`
				VaaKeys                []any  `bson:"vaaKeys" json:"vaaKeys"`
				TargetChainID          string `bson:"targetChainId" json:"targetChainId"`
			} `bson:"deliveryInstructionsPrintable" json:"deliveryInstructionsPrintable"`
			WalletAddress              string  `bson:"walletAddress" json:"walletAddress"`
			GasUsed                    int     `bson:"gasUsed" json:"gasUsed"`
			GasPrice                   string  `bson:"gasPrice" json:"gasPrice"`
			ReceiverValue              string  `bson:"receiverValue" json:"receiverValue"`
			MaxRefundUsd               float64 `bson:"maxRefundUsd" json:"maxRefundUsd"`
			GasPriceEstimate           string  `bson:"gasPriceEstimate" json:"gasPriceEstimate"`
			TransactionDidSubmit       bool    `bson:"transactionDidSubmit" json:"transactionDidSubmit"`
			EstimatedTransactionFee    string  `bson:"estimatedTransactionFee" json:"estimatedTransactionFee"`
			TransactionSubmitTimeStart int64   `bson:"transactionSubmitTimeStart" json:"transactionSubmitTimeStart"`
			TransactionSubmitTimeEnd   int64   `bson:"transactionSubmitTimeEnd" json:"transactionSubmitTimeEnd"`
			ResultLogDidParse          bool    `bson:"resultLogDidParse" json:"resultLogDidParse"`
			ChainID                    int     `bson:"chainId" json:"chainId"`
			ReceiverValueUsd           float64 `bson:"receiverValueUsd" json:"receiverValueUsd"`
			WalletBalanceBefore        string  `bson:"walletBalanceBefore" json:"walletBalanceBefore"`
		} `bson:"deliveryRecord"`
		RawVaaHex            string `bson:"rawVaaHex" json:"rawVaaHex"`
		PayloadType          int    `bson:"payloadType" json:"payloadType"`
		MaxAttempts          int    `bson:"maxAttempts" json:"maxAttempts"`
		DidError             bool   `bson:"didError" json:"didError"`
		ExecutionEndTime     int64  `bson:"executionEndTime" json:"executionEndTime"`
		EmitterAddress       string `bson:"emitterAddress" json:"emitterAddress"`
		DidSubmitTransaction bool   `bson:"didSubmitTransaction" json:"didSubmitTransaction"`
		Sequence             string `bson:"sequence" json:"sequence"`
	} `bson:"metadata" json:"metadata"`
	Sequence       string     `bson:"sequence" json:"sequence"`
	Vaa            string     `bson:"vaa" json:"vaa"`
	FromTxHash     string     `bson:"fromTxHash" json:"fromTxHash"`
	MaxAttempts    int        `bson:"maxAttempts" json:"maxAttempts"`
	AddedTimes     int        `bson:"addedTimes" json:"addedTimes"`
	ErrorMessage   any        `bson:"errorMessage" json:"errorMessage"`
	EmitterChain   int        `bson:"emitterChain" json:"emitterChain"`
	EmitterAddress string     `bson:"emitterAddress" json:"emitterAddress"`
	FailedAt       *time.Time `bson:"failedAt" json:"failedAt"`
}

type RelayDoc struct {
	ID     string    `bson:"_id" json:"id"`
	Data   RelayData `bson:"data" json:"data"`
	Event  string    `bson:"event" json:"event"`
	Origin string    `bson:"origin" json:"origin"`
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
