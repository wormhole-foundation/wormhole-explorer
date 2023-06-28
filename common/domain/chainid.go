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
		sdk.ChainIDKarura,
		sdk.ChainIDAcala,
		sdk.ChainIDKlaytn,
		sdk.ChainIDCelo,
		sdk.ChainIDMoonbeam,
		sdk.ChainIDArbitrum,
		sdk.ChainIDOptimism:

		return "0x" + hex.EncodeToString(addressBytes[12:]), nil

	case sdk.ChainIDTerra:
		aligned, err := bech32.ConvertBits(addressBytes[12:], 8, 5, true)
		if err != nil {
			return "", fmt.Errorf("encoding terra bech32 failed: %w", err)
		}
		return bech32.Encode("terra", aligned)

	case sdk.ChainIDTerra2:
		aligned, err := bech32.ConvertBits(addressBytes, 8, 5, true)
		if err != nil {
			return "", fmt.Errorf("encoding terra2 bech32 failed: %w", err)
		}
		return bech32.Encode("terra", aligned)

	case sdk.ChainIDInjective:
		aligned, err := bech32.ConvertBits(addressBytes[12:], 8, 5, true)
		if err != nil {
			return "", fmt.Errorf("encoding injective bech32 failed: %w", err)
		}
		return bech32.Encode("inj", aligned)

	case sdk.ChainIDXpla:
		aligned, err := bech32.ConvertBits(addressBytes, 8, 5, true)
		if err != nil {
			return "", fmt.Errorf("encoding xpla bech32 failed: %w", err)
		}
		return bech32.Encode("xpla", aligned)

	case sdk.ChainIDSei:
		aligned, err := bech32.ConvertBits(addressBytes, 8, 5, true)
		if err != nil {
			return "", fmt.Errorf("encoding sei bech32 failed: %w", err)
		}
		return bech32.Encode("sei", aligned)

	case sdk.ChainIDAlgorand:

		var addr algorand_types.Address
		if len(addr) != len(addressBytes) {
			return "", fmt.Errorf("expected Algorand address to be %d bytes long, but got: %d", len(addr), len(addressBytes))
		}
		copy(addr[:], addressBytes[:])

		return addr.String(), nil

	case sdk.ChainIDNear:
		if nativeAddress, ok := nearMappings[address]; ok {
			return nativeAddress, nil
		} else {
			return "", fmt.Errorf(`no mapping found for NEAR emitter address "%s"`, address)
		}

	case sdk.ChainIDSui:
		if nativeAddress, ok := suiMappings[address]; ok {
			return nativeAddress, nil
		} else {
			return "", fmt.Errorf(`no mapping found for Sui emitter address "%s"`, address)
		}

	case sdk.ChainIDAptos:
		if nativeAddress, ok := aptosMappings[address]; ok {
			return nativeAddress, nil
		} else {
			return "", fmt.Errorf(`no mapping found for Aptos emitter address "%s"`, address)
		}

	default:
		return "", fmt.Errorf("can't translate emitter address: ChainID=%d not supported", chainID)
	}
}

var nearMappings = map[string]string{
	"148410499d3fcda4dcfd68a1ebfcdddda16ab28326448d4aae4d2f0465cdfcb7": "contract.portalbridge.near",
}

var suiMappings = map[string]string{
	"ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5": "0xc57508ee0d4595e5a8728974a4a93a787d38f339757230d441e895422c07aba9",
}

var aptosMappings = map[string]string{
	"0000000000000000000000000000000000000000000000000000000000000001": "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
}
