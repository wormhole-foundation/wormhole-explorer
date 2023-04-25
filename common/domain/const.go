package domain

// p2p network constants.
const (
	P2pMainNet = "mainnet"
	P2pTestNet = "testnet"
	P2pDevNet  = "devnet"
)

const AppIdPortalTokenBridge = "PORTAL_TOKEN_BRIDGE"

// SourceTxStatus is meant to be a user-facing enum that describes the status of the source transaction.
type SourceTxStatus string

const (
	// SourceTxStatusChainNotSupported indicates that the processing failed due to the chain ID not being supported.
	//
	// (i.e.: there is no adapter for that chain yet)
	SourceTxStatusChainNotSupported SourceTxStatus = "chainNotSupported"

	// SourceTxStatusInternalError represents an internal, unspecified error.
	SourceTxStatusInternalError SourceTxStatus = "internalError"

	// SourceTxStatusConfirmed indicates that the transaciton has been processed successfully.
	SourceTxStatusConfirmed SourceTxStatus = "confirmed"
)
