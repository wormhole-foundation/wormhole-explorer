package domain

import (
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"

	algorand_types "github.com/algorand/go-algorand-sdk/types"
	"github.com/cosmos/btcutil/bech32"
	"github.com/mr-tron/base58"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

var (
	// nearKnownEmitters maps NEAR emitter addresses to NEAR accounts.
	nearKnownEmitters = map[string]string{
		"148410499d3fcda4dcfd68a1ebfcdddda16ab28326448d4aae4d2f0465cdfcb7": "contract.portalbridge.near",
	}

	// suiKnownEmitters maps Sui emitter addresses to Sui accounts.
	suiKnownEmitters = map[string]string{
		"ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5": "0xc57508ee0d4595e5a8728974a4a93a787d38f339757230d441e895422c07aba9",
	}

	// aptosKnownEmitters maps Aptos emitter addresses to Aptos accounts.
	aptosKnownEmitters = map[string]string{
		// Token Bridge
		"0000000000000000000000000000000000000000000000000000000000000001": "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
		// NFT Bridge
		"0000000000000000000000000000000000000000000000000000000000000005": "0x1bdffae984043833ed7fe223f7af7a3f8902d04129b14f801823e64827da7130",
	}
)

var allChainIDs = make(map[sdk.ChainID]bool)

func init() {
	for _, chainID := range sdk.GetAllNetworkIDs() {
		allChainIDs[chainID] = true
	}
}

// ChainIdIsValid returns true if and only if the given chain ID exists.
func ChainIdIsValid(id sdk.ChainID) bool {
	_, exists := allChainIDs[id]
	return exists
}

// GetSupportedChainIDs returns a map of all supported chain IDs to their respective names.
func GetSupportedChainIDs() map[sdk.ChainID]string {
	chainIDs := sdk.GetAllNetworkIDs()
	supportedChaindIDs := make(map[sdk.ChainID]string, len(chainIDs))
	for _, chainID := range chainIDs {
		supportedChaindIDs[chainID] = chainID.String()
	}
	return supportedChaindIDs
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

	// Solana emitter addresses use base58 encoding.
	case sdk.ChainIDSolana:
		return base58.Encode(addressBytes), nil

	// EVM chains use the classic hex, 0x-prefixed encoding.
	// Also, Karura and Acala support EVM-compatible addresses, so they're handled here as well.
	case sdk.ChainIDEthereum,
		sdk.ChainIDBase,
		sdk.ChainIDBSC,
		sdk.ChainIDPolygon,
		sdk.ChainIDPolygonSepolia,
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
		sdk.ChainIDOptimism,
		sdk.ChainIDSepolia,
		sdk.ChainIDArbitrumSepolia,
		sdk.ChainIDBaseSepolia,
		sdk.ChainIDOptimismSepolia,
		sdk.ChainIDHolesky,
		sdk.ChainIDWormchain,
		sdk.ChainIDScroll,
		sdk.ChainIDBlast:

		return "0x" + hex.EncodeToString(addressBytes[12:]), nil

	// Terra addresses use bench32 encoding
	case sdk.ChainIDTerra:
		return encodeBech32("terra", addressBytes[12:])

	// Terra2 addresses use bench32 encoding
	case sdk.ChainIDTerra2:
		return encodeBech32("terra", addressBytes)

	// Injective addresses use bench32 encoding
	case sdk.ChainIDInjective:
		return encodeBech32("inj", addressBytes[12:])

	// Xpla addresses use bench32 encoding
	case sdk.ChainIDXpla:
		return encodeBech32("xpla", addressBytes)

	// Sei addresses use bench32 encoding
	case sdk.ChainIDSei:
		return encodeBech32("sei", addressBytes)

	// Algorand addresses use base32 encoding with a trailing checksum.
	// We're using the SDK to handle the checksum logic.
	case sdk.ChainIDAlgorand:

		var addr algorand_types.Address
		if len(addr) != len(addressBytes) {
			return "", fmt.Errorf("expected Algorand address to be %d bytes long, but got: %d", len(addr), len(addressBytes))
		}
		copy(addr[:], addressBytes[:])

		return addr.String(), nil

	// Near addresses are arbitrary-length strings. The emitter is the sha256 digest of the program address string.
	//
	// We're using a hashmap of known emitters to avoid querying external APIs.
	case sdk.ChainIDNear:
		if nativeAddress, ok := nearKnownEmitters[address]; ok {
			return nativeAddress, nil
		} else {
			return "", fmt.Errorf(`no mapping found for NEAR emitter address "%s"`, address)
		}

	// For Sui emitters, an emitter capacity is taken from the core bridge. The capability object ID is used.
	//
	// We're using a hashmap of known emitters to avoid querying the contract's state.
	case sdk.ChainIDSui:
		if nativeAddress, ok := suiKnownEmitters[address]; ok {
			return nativeAddress, nil
		} else {
			return "", fmt.Errorf(`no mapping found for Sui emitter address "%s"`, address)
		}

	// For Aptos, an emitter capability is taken from the core bridge. The capability object ID is used.
	// The core bridge generates capabilities in a sequence and the capability object ID is its index in the sequence.
	//
	// We're using a hashmap of known emitters to avoid querying the contract's state.
	case sdk.ChainIDAptos:
		if nativeAddress, ok := aptosKnownEmitters[address]; ok {
			return nativeAddress, nil
		} else {
			return "", fmt.Errorf(`no mapping found for Aptos emitter address "%s"`, address)
		}

	default:
		return "", fmt.Errorf("can't translate emitter address: ChainID=%d not supported", chainID)
	}
}

func NormalizeTxHashByChainId(chainID sdk.ChainID, txHash string) string {
	switch chainID {
	case sdk.ChainIDEthereum,
		sdk.ChainIDBase,
		sdk.ChainIDBSC,
		sdk.ChainIDPolygon,
		sdk.ChainIDPolygonSepolia,
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
		sdk.ChainIDOptimism,
		sdk.ChainIDSepolia,
		sdk.ChainIDArbitrumSepolia,
		sdk.ChainIDBaseSepolia,
		sdk.ChainIDOptimismSepolia,
		sdk.ChainIDHolesky:
		lowerTxHash := strings.ToLower(txHash)
		return utils.Remove0x(lowerTxHash)
	default:
		return txHash
	}
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
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDSei:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDWormchain:
		//TODO: check if this is correct
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDScroll:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDBlast:
		return hex.EncodeToString(txHash), nil
	case sdk.ChainIDSepolia,
		sdk.ChainIDArbitrumSepolia,
		sdk.ChainIDBaseSepolia,
		sdk.ChainIDOptimismSepolia,
		sdk.ChainIDHolesky,
		sdk.ChainIDPolygonSepolia:
		return hex.EncodeToString(txHash), nil
	default:
		return hex.EncodeToString(txHash), fmt.Errorf("unknown chain id: %d", chainID)
	}
}

// DecodeNativeAddressToHex decodes a native address to hex.
func DecodeNativeAddressToHex(chainID sdk.ChainID, address string) (string, error) {

	// Translation rules are based on the chain ID
	switch chainID {

	// Solana emitter addresses use base58 encoding.
	case sdk.ChainIDSolana:
		addr, err := base58.Decode(address)
		if err != nil {
			return "", fmt.Errorf("base58 decoding failed: %w", err)
		}
		return hex.EncodeToString(addr), nil

	// EVM chains use the classic hex, 0x-prefixed encoding.
	// Also, Karura and Acala support EVM-compatible addresses, so they're handled here as well.
	case sdk.ChainIDEthereum,
		sdk.ChainIDBase,
		sdk.ChainIDBSC,
		sdk.ChainIDPolygon,
		sdk.ChainIDPolygonSepolia,
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
		sdk.ChainIDOptimism,
		sdk.ChainIDSepolia,
		sdk.ChainIDArbitrumSepolia,
		sdk.ChainIDBaseSepolia,
		sdk.ChainIDOptimismSepolia,
		sdk.ChainIDHolesky,
		sdk.ChainIDWormchain,
		sdk.ChainIDScroll,
		sdk.ChainIDBlast:

		return address, nil

	// Terra addresses use bench32 encoding
	case sdk.ChainIDTerra:
		return decodeBech32("terra", address)

	// Terra2 addresses use bench32 encoding
	case sdk.ChainIDTerra2:
		return decodeBech32("terra", address)

	// Injective addresses use bench32 encoding
	case sdk.ChainIDInjective:
		return decodeBech32("inj", address)

	// Sui addresses use hex encoding
	case sdk.ChainIDSui:
		return address, nil

	// Aptos addresses use hex encoding
	case sdk.ChainIDAptos:
		return address, nil

	// Xpla addresses use bench32 encoding
	case sdk.ChainIDXpla:
		return decodeBech32("xpla", address)

	// Sei addresses use bench32 encoding
	case sdk.ChainIDSei:
		return decodeBech32("sei", address)

	// Algorand addresses use base32 encoding with a trailing checksum.
	// We're using the SDK to handle the checksum logic.
	case sdk.ChainIDAlgorand:
		addr, err := algorand_types.DecodeAddress(address)
		if err != nil {
			return "", fmt.Errorf("algorand decoding failed: %w", err)
		}
		return hex.EncodeToString(addr[:]), nil

	default:
		return "", fmt.Errorf("can't translate emitter address: ChainID=%d not supported", chainID)
	}
}

// decodeBech32 is a helper function to decode a bech32 addresses.
func decodeBech32(h, address string) (string, error) {

	hrp, decoded, err := bech32.Decode(address, bech32.MaxLengthBIP173)
	if err != nil {
		return "", fmt.Errorf("bech32 decoding failed: %w", err)
	}
	if hrp != h {
		return "", fmt.Errorf("bech32 decoding failed, invalid prefix: %s", hrp)
	}

	return hex.EncodeToString(decoded), nil
}

// encodeBech32 is a helper function to encode a bech32 addresses.
func encodeBech32(hrp string, data []byte) (string, error) {

	aligned, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("bech32 encoding failed: %w", err)
	}

	return bech32.Encode(hrp, aligned)
}
