package domain

import (
	"encoding/base32"
	"encoding/hex"
	"fmt"

	algorand_types "github.com/algorand/go-algorand-sdk/types"
	"github.com/cosmos/btcutil/bech32"
	"github.com/mr-tron/base58"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// GetSupportedChainIDs returns a map of all supported chain IDs to their respective names.
func GetSupportedChainIDs() map[sdk.ChainID]string {
	chainIDs := sdk.GetAllNetworkIDs()
	supportedChaindIDs := make(map[sdk.ChainID]string, len(chainIDs))
	for _, chainID := range chainIDs {
		supportedChaindIDs[chainID] = chainID.String()
	}
	return supportedChaindIDs
}

// EncodeTrxHashByChainID encodes the transaction hash by chain id with different encoding methods.
func EncodeTrxHashByChainID(chainID sdk.ChainID, txHash []byte) (string, error) {
	switch chainID {
	case sdk.ChainIDSolana:
		return base58.Encode(txHash), nil
	case sdk.ChainIDEthereum:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDTerra:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDBSC:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDPolygon:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDAvalanche:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDOasis:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDAlgorand:
		return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(txHash), nil
	case sdk.ChainIDAurora:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDFantom:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDKarura:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDAcala:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDKlaytn:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDCelo:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDNear:
		return base58.Encode(txHash), nil
	case sdk.ChainIDMoonbeam:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDNeon:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDTerra2:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDInjective:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDSui:
		return base58.Encode(txHash), nil
	case sdk.ChainIDAptos:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDArbitrum:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDOptimism:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDXpla:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDBtc:
		//TODO: check if this is correct
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDBase:
		//TODO: check if this is correct
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDSei:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDWormchain:
		//TODO: check if this is correct
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDSepolia:
		return hex.EncodeToString(txHash), nil
	default:
		return hex.EncodeToString(txHash), fmt.Errorf("unknown chain id: %d", chainID)
	}
}

// TranslateEmitterAddress converts an emitter address into the corresponding native address for the given chain.
func TranslateEmitterAddress(chainID sdk.ChainID, address string) (string, error) {

	// Decode the address from hex
	addressBytes, err := hex.DecodeString(address)
	if err != nil {
		return "", fmt.Errorf(`failed to decode emitter address "%s" from hex: %w`, address, err)
	}
	if len(addressBytes) != 32 {
		return "", fmt.Errorf("expected emitter address length to be 32: %s", address)
	}

	// Translation rules are based on the chain ID
	switch chainID {

	case sdk.ChainIDSolana:
		return base58.Encode(addressBytes), nil

	case sdk.ChainIDEthereum,
		sdk.ChainIDBSC,
		sdk.ChainIDPolygon,
		sdk.ChainIDAvalanche,
		sdk.ChainIDOasis,
		sdk.ChainIDAurora,
		sdk.ChainIDFantom,
		sdk.ChainIDKarura:

		return "0x" + hex.EncodeToString(addressBytes[12:]), nil

	case sdk.ChainIDTerra:
		aligned, err := bech32.ConvertBits(addressBytes[12:], 8, 5, true)
		if err != nil {
			return "", fmt.Errorf("encoding bech32 failed: %w", err)
		}
		return bech32.Encode("terra", aligned)

	case sdk.ChainIDAlgorand:

		var addr algorand_types.Address
		if len(addr) != len(addressBytes) {
			return "", fmt.Errorf("expected Algorand address to be %d bytes long, but got: %d", len(addr), len(addressBytes))
		}
		copy(addr[:], addressBytes[:])

		return addr.String(), nil

	default:
		return "", fmt.Errorf("can't translate emitter address: ChainID=%d not supported", chainID)
	}
}
