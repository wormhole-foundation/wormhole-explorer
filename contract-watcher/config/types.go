package config

import "github.com/wormhole-foundation/wormhole/sdk/vaa"

const (
	//Method names for wormhole token bridge contract.
	MethodCompleteTransfer     = "completeTransfer"
	MethodWrapAndTransfer      = "wrapAndTransfer"
	MethodTransferTokens       = "transferTokens"
	MethodAttestToken          = "attestToken"
	MethodCompleteAndUnwrapETH = "completeAndUnwrapETH"
	MethodCreateWrapped        = "createWrapped"
	MethodUpdateWrapped        = "updateWrapped"
	MethodUnkown               = "unknown"
	//Method name for wormhole connect wrapped contract.
	MetehodCompleteTransferWithRelay = "completeTransferWithRelay"

	//Method name for wormhole tBTC gateway
	MethodReceiveTbtc = "receiveTbtc"

	//Method name for Portico contract
	MethodReceiveMessageAndSwap = "receiveMessageAndSwap"

	//Method ids for wormhole token bridge contract
	MethodIDCompleteTransfer     = "0xc6878519"
	MethodIDWrapAndTransfer      = "0x9981509f"
	MethodIDTransferTokens       = "0x0f5287b0"
	MethodIDAttestToken          = "0xc48fa115"
	MethodIDCompleteAndUnwrapETH = "0xff200cde"
	MethodIDCreateWrapped        = "0xe8059810"
	MethodIDUpdateWrapped        = "0xf768441f"
	//Method id for wormhole connect wrapped contract.
	MetehodIDCompleteTransferWithRelay = "0x2f25e25f"

	//Method id for wormhole tBTC gateway
	MethodIDReceiveTbtc = "0x5d21a596"

	//Method id for Portico contract
	MethodIDReceiveMessageAndSwap = "0x3d528f35"
)

type WatcherBlockchain struct {
	ChainID      vaa.ChainID
	Name         string
	Address      string
	SizeBlocks   uint8
	WaitSeconds  uint16
	InitialBlock int64
}

type WatcherBlockchainAddresses struct {
	ChainID     vaa.ChainID
	Name        string
	SizeBlocks  uint8
	WaitSeconds uint16
	// Initial block indicates for the supported contracts, the oldest block from which to start processing.
	InitialBlock     int64
	MethodsByAddress map[string][]BlockchainMethod
}

type BlockchainMethod struct {
	ID   string
	Name string
}
