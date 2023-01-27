// package middleare contains all the middleware function to use in the API.
package middleware

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// ExtractChainID get chain parameter from route path.
func ExtractChainID(c *fiber.Ctx, l *zap.Logger) (vaa.ChainID, error) {
	chain, err := c.ParamsInt("chain")
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to get chain parameter", zap.Error(err), zap.Int("chain", chain),
			zap.String("requestID", requestID))

		return vaa.ChainIDUnset, response.NewInvalidParamError(c, "WRONG CHAIN ID", errors.WithStack(err))
	}
	return vaa.ChainID(chain), nil
}

// ExtractEmitterAddr get emitter parameter from route path.
func ExtractEmitterAddr(c *fiber.Ctx, l *zap.Logger) (*vaa.Address, error) {
	emitterStr := c.Params("emitter")
	emitter, err := vaa.StringToAddress(emitterStr)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to covert emitter to address", zap.Error(err), zap.String("emitterStr", emitterStr),
			zap.String("requestID", requestID))
		return nil, response.NewInvalidParamError(c, "MALFORMED EMITTER_ADDR", errors.WithStack(err))
	}
	return &emitter, nil
}

// ExtractSequence get sequence parameter from route path.
func ExtractSequence(c *fiber.Ctx, l *zap.Logger) (uint64, error) {
	sequence := c.Params("sequence")
	seq, err := strconv.ParseUint(sequence, 10, 64)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to get sequence parameter", zap.Error(err), zap.String("sequence", sequence),
			zap.String("requestID", requestID))
		return 0, response.NewInvalidParamError(c, "MALFORMED SEQUENCE NUMBER", errors.WithStack(err))
	}
	return seq, nil
}

// ExtractGuardianAddress get guardian address from route path.
func ExtractGuardianAddress(c *fiber.Ctx, l *zap.Logger) (string, error) {
	//TODO: check guardianAddress [vaa.StringToAddress(emitterStr)]
	guardianAddress := c.Params("guardian_address")
	if guardianAddress == "" {
		return "", response.NewInvalidParamError(c, "MALFORMED GUARDIAN ADDR", nil)
	}
	return guardianAddress, nil
}

// ExtractVAAParams get VAA chain, address from route path.
func ExtractVAAChainIDEmitter(c *fiber.Ctx, l *zap.Logger) (vaa.ChainID, *vaa.Address, error) {
	chainID, err := ExtractChainID(c, l)
	if err != nil {
		return vaa.ChainIDUnset, nil, err
	}
	address, err := ExtractEmitterAddr(c, l)
	if err != nil {
		return chainID, nil, err
	}
	return chainID, address, nil
}

// ExtractVAAParams get VAAA chain, address and sequence from route path.
func ExtractVAAParams(c *fiber.Ctx, l *zap.Logger) (vaa.ChainID, *vaa.Address, uint64, error) {
	chainID, err := ExtractChainID(c, l)
	if err != nil {
		return vaa.ChainIDUnset, nil, 0, err
	}
	address, err := ExtractEmitterAddr(c, l)
	if err != nil {
		return chainID, nil, 0, err
	}
	seq, err := ExtractSequence(c, l)
	if err != nil {
		return chainID, address, 0, err
	}
	return chainID, address, seq, nil
}

// ExtractObservationSigner get signer from route path.
func ExtractObservationSigner(c *fiber.Ctx, l *zap.Logger) (*vaa.Address, error) {
	signer := c.Params("signer")
	signerAddr, err := vaa.StringToAddress(signer)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to covert signer to address", zap.Error(err), zap.String("signer", signer),
			zap.String("requestID", requestID))
		return nil, response.NewInvalidParamError(c, "MALFORMED SIGNER", errors.WithStack(err))
	}
	return &signerAddr, nil
}

// ExtractObservationHash get a hash from route path.
func ExtractObservationHash(c *fiber.Ctx, l *zap.Logger) (string, error) {
	hash := c.Params("hash")
	if hash == "" {
		return "", response.NewInvalidParamError(c, "MALFORMED HASH", nil)
	}
	return hash, nil
}

// GetTxHash get txHash parameter from query param.
func GetTxHash(c *fiber.Ctx, l *zap.Logger) (*vaa.Address, error) {
	txHash := c.Query("txHash")
	if txHash == "" {
		return nil, nil
	}
	txHashAddr, err := vaa.StringToAddress(txHash)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to covert txHash to address", zap.Error(err), zap.String("txHash", txHash),
			zap.String("requestID", requestID))
		return nil, response.NewInvalidParamError(c, "MALFORMED TX HASH", errors.WithStack(err))
	}
	return &txHashAddr, nil
}

// ExtractParsedPayload get parsedPayload query parameter.
func ExtractParsedPayload(c *fiber.Ctx, l *zap.Logger) (bool, error) {
	parsedPayloadStr := c.Query("parsedPayload", "false")
	parsedPayload, err := strconv.ParseBool(parsedPayloadStr)
	if err != nil {
		return false, response.NewInvalidQueryParamError(c, "INVALID <parsedPayload> QUERY PARAMETER", errors.WithStack(err))
	}
	return parsedPayload, nil
}
