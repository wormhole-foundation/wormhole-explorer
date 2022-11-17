// package middleare contains all the middleware nec
package middleware

import (
	"strconv"

	"github.com/certusone/wormhole/node/pkg/vaa"
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/errs"
)

var (
	ErrMalformedChain             = errs.NewParamError("WRONG CHAIN ID")
	ErrMalformedAddr              = errs.NewParamError("MALFORMED EMITTER_ADDR")
	ErrMalformedSequence          = errs.NewParamError("MALFORMED SEQUENCE NUMBER")
	ErrMalFormedGuardianAddress   = errs.NewParamError("MALFORMED GUARDIAN ADDR")
	ErrMalFormedObservationSigner = errs.NewParamError("MALFORMED SIGNER")
	ErrMalFormedObservationHash   = errs.NewParamError("MALFORMED HASH")
)

// ExtractChainID get chain parameter from route path.
func ExtractChainID(c *fiber.Ctx) (vaa.ChainID, error) {
	chain, err := c.ParamsInt("chain")
	if err != nil {
		return vaa.ChainIDUnset, ErrMalformedChain
	}
	return vaa.ChainID(chain), nil
}

// ExtractEmitterAddr get emitter parameter from route path.
func ExtractEmitterAddr(c *fiber.Ctx) (*vaa.Address, error) {
	emitterStr := c.Params("emitter")
	emitter, err := vaa.StringToAddress(emitterStr)
	if err != nil {
		return nil, ErrMalformedAddr
	}
	return &emitter, nil
}

// ExtractSequence get sequence parameter from route path.
func ExtractSequence(c *fiber.Ctx) (uint64, error) {
	sequence := c.Params("sequence")
	seq, err := strconv.ParseUint(sequence, 10, 64)
	if err != nil {
		return 0, ErrMalformedSequence
	}
	return seq, nil
}

// ExtractGuardianAddress get guardian address from route path.
func ExtractGuardianAddress(c *fiber.Ctx) (string, error) {
	//TODO: check guardianAddress [vaa.StringToAddress(emitterStr)]
	guardianAddress := c.Params("guardian_address")
	if guardianAddress == "" {
		return "", ErrMalFormedGuardianAddress
	}
	return guardianAddress, nil
}

// ExtractVAAParams get VAA chain, address from route path.
func ExtractVAAChainIDEmitter(c *fiber.Ctx) (vaa.ChainID, *vaa.Address, error) {
	chainID, err := ExtractChainID(c)
	if err != nil {
		return vaa.ChainIDUnset, nil, err
	}
	address, err := ExtractEmitterAddr(c)
	if err != nil {
		return chainID, nil, err
	}
	return chainID, address, nil
}

// ExtractVAAParams get VAAA chain, address and sequence from route path.
func ExtractVAAParams(c *fiber.Ctx) (vaa.ChainID, *vaa.Address, uint64, error) {
	chainID, err := ExtractChainID(c)
	if err != nil {
		return vaa.ChainIDUnset, nil, 0, err
	}
	address, err := ExtractEmitterAddr(c)
	if err != nil {
		return chainID, nil, 0, err
	}
	seq, err := ExtractSequence(c)
	if err != nil {
		return chainID, address, 0, err
	}
	return chainID, address, seq, nil
}

// ExtractObservationSigner get signer from route path.
func ExtractObservationSigner(c *fiber.Ctx) (*vaa.Address, error) {
	signer := c.Params("signer")
	signerAddr, err := vaa.StringToAddress(signer)
	if err != nil {
		return nil, ErrMalFormedObservationSigner
	}
	return &signerAddr, nil
}

// ExtractObservationHash get a hash from route path.
func ExtractObservationHash(c *fiber.Ctx) (string, error) {
	hash := c.Params("hash")
	if hash == "" {
		return "", ErrMalFormedObservationHash
	}
	return hash, nil
}
