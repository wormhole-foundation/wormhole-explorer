package domain

// p2p network constants.
const (
	P2pMainNet = "mainnet"
	P2pTestNet = "testnet"
	P2pDevNet  = "devnet"
)

type TxStatus uint

const (
	TxStatusChainNotSupported  TxStatus = 0
	TxStatusFailedToProcess    TxStatus = 1
	TxStatusConfirmed          TxStatus = 2
)
