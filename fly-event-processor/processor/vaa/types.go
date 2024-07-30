package vaa

import (
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"golang.org/x/net/context"
)

// Params is a struct to store the parameters for the processor.
type Params struct {
	TrackID string
	VaaID   string
	ChainID sdk.ChainID
}

// ProcessorFunc is a function to process vaa message.
type ProcessorFunc func(context.Context, *Params) error

// getFinalityTimeByChainID returns the finality time for each chain.
func getFinalityTimeByChainID(chainID sdk.ChainID) time.Duration {
	// Time to finalize for each chain.
	// ref: https://docs.wormhole.com/wormhole/reference/constants
	switch chainID {
	case sdk.ChainIDSolana:
		return 14 * time.Second
	case sdk.ChainIDEthereum:
		return 975 * time.Second
	case sdk.ChainIDTerra:
		return 6 * time.Second
	case sdk.ChainIDBSC:
		return 48 * time.Second
	case sdk.ChainIDPolygon:
		return 66 * time.Second
	case sdk.ChainIDAvalanche:
		return 2 * time.Second
	case sdk.ChainIDOasis:
		return 12 * time.Second
	case sdk.ChainIDAlgorand:
		return 4 * time.Second
	case sdk.ChainIDFantom:
		return 5 * time.Second
	case sdk.ChainIDKarura:
		return 24 * time.Second
	case sdk.ChainIDAcala:
		return 24 * time.Second
	case sdk.ChainIDKlaytn:
		return 1 * time.Second
	case sdk.ChainIDCelo:
		return 10 * time.Second
	case sdk.ChainIDNear:
		return 2 * time.Second
	case sdk.ChainIDMoonbeam:
		return 24 * time.Second
	case sdk.ChainIDTerra2:
		return 6 * time.Second
	case sdk.ChainIDInjective:
		return 3 * time.Second
	case sdk.ChainIDSui:
		return 3 * time.Second
	case sdk.ChainIDAptos:
		return 4 * time.Second
	case sdk.ChainIDArbitrum:
		return 1066 * time.Second
	case sdk.ChainIDOptimism:
		return 1026 * time.Second
	case sdk.ChainIDXpla:
		return 5 * time.Second
	case sdk.ChainIDBase:
		return 1026 * time.Second
	case sdk.ChainIDSei:
		return 1 * time.Second
	case sdk.ChainIDScroll:
		return 1200 * time.Second
	case sdk.ChainIDMantle:
		return 1200 * time.Second
	case sdk.ChainIDBlast:
		return 1200 * time.Second
	case sdk.ChainIDXLayer:
		return 1200 * time.Second
	case sdk.ChainIDBerachain:
		return 5 * time.Second
	case sdk.ChainIDWormchain:
		return 5 * time.Second
	case sdk.ChainIDSepolia:
		return 975 * time.Second
	case sdk.ChainIDArbitrumSepolia:
		return 1066 * time.Second
	case sdk.ChainIDBaseSepolia:
		return 1026 * time.Second
	case sdk.ChainIDOptimismSepolia:
		return 1026 * time.Second
	case sdk.ChainIDHolesky:
		return 975 * time.Second
	default:
		// The default value is the max finality time.
		return 1066 * time.Second
	}
}
