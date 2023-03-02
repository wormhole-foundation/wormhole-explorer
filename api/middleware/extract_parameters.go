// package middleare contains all the middleware function to use in the API.
package middleware

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole-explorer/api/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// ExtractChainID get chain parameter from route path.
func ExtractChainID(c *fiber.Ctx, l *zap.Logger) (vaa.ChainID, error) {

	chain, err := c.ParamsInt("chain")
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to get chain parameter",
			zap.Error(err),
			zap.Int("chain", chain),
			zap.String("requestID", requestID),
		)

		return vaa.ChainIDUnset, response.NewInvalidParamError(c, "WRONG CHAIN ID", errors.WithStack(err))
	}

	return vaa.ChainID(chain), nil
}

// ExtractEmitterAddr get emitter parameter from route path.
func ExtractEmitterAddr(c *fiber.Ctx, l *zap.Logger) (*types.Address, error) {

	emitterStr := c.Params("emitter")

	emitter, err := types.StringToAddress(emitterStr)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to convert emitter to address",
			zap.Error(err),
			zap.String("emitterStr", emitterStr),
			zap.String("requestID", requestID),
		)
		return nil, response.NewInvalidParamError(c, "MALFORMED EMITTER_ADDR", errors.WithStack(err))
	}

	return emitter, nil
}

// ExtractSequence get sequence parameter from route path.
func ExtractSequence(c *fiber.Ctx, l *zap.Logger) (uint64, error) {

	sequence := c.Params("sequence")

	seq, err := strconv.ParseUint(sequence, 10, 64)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to get sequence parameter",
			zap.Error(err),
			zap.String("sequence", sequence),
			zap.String("requestID", requestID),
		)
		return 0, response.NewInvalidParamError(c, "MALFORMED SEQUENCE NUMBER", errors.WithStack(err))
	}

	return seq, nil
}

// ExtractGuardianAddress get guardian address from route path.
func ExtractGuardianAddress(c *fiber.Ctx, l *zap.Logger) (*types.Address, error) {

	// read the address from query params
	tmp := c.Params("guardian_address")
	if tmp == "" {
		return nil, response.NewInvalidParamError(c, "MALFORMED GUARDIAN ADDR", nil)
	}

	// validate the address
	guardianAddress, err := types.StringToAddress(tmp)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to decode guardian address",
			zap.Error(err),
			zap.String("requestID", requestID),
		)
		return nil, response.NewInvalidParamError(c, "MALFORMED GUARDIAN ADDR", errors.WithStack(err))
	}

	return guardianAddress, nil
}

// ExtractVAAParams get VAA chain, address from route path.
func ExtractVAAChainIDEmitter(c *fiber.Ctx, l *zap.Logger) (vaa.ChainID, *types.Address, error) {

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
func ExtractVAAParams(c *fiber.Ctx, l *zap.Logger) (vaa.ChainID, *types.Address, uint64, error) {

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
		l.Error("failed to covert signer to address",
			zap.Error(err),
			zap.String("signer", signer),
			zap.String("requestID", requestID),
		)
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
		l.Error("failed to covert txHash to address",
			zap.Error(err),
			zap.String("txHash", txHash),
			zap.String("requestID", requestID),
		)
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

func ExtractAppId(c *fiber.Ctx, l *zap.Logger) string {
	return c.Query("appId")
}

func ExtractTimeSpan(c *fiber.Ctx, l *zap.Logger) (string, error) {
	// get the timeSpan from query params
	timeSpanStr := c.Query("timeSpan", "all")
	if timeSpanStr == "all" {
		return timeSpanStr, nil
	}

	// validate the timeSpan
	if !isValidTimeSpan(timeSpanStr) {
		return "", response.NewInvalidQueryParamError(c, "INVALID <timeSpan> QUERY PARAMETER", nil)
	}
	return timeSpanStr, nil
}

// isValidTimeSpan check if the timeSpan is valid
func isValidTimeSpan(timeSpan string) bool {
	return regexp.MustCompile(`^all$|^\d+[mhdwy]$|^\dmo$`).MatchString(timeSpan)
}

func ExtractSampleRate(c *fiber.Ctx, l *zap.Logger) (string, error) {
	// get the sampleRate from query params
	sampleRateStr := c.Query("sampleRate", "1y")
	if sampleRateStr == "1y" {
		return sampleRateStr, nil
	}
	// validate the sampleRate
	if !isValidSampleRate(sampleRateStr) {
		return "", response.NewInvalidQueryParamError(c, "INVALID <sampleRate> QUERY PARAMETER", nil)
	}
	return sampleRateStr, nil
}

func isValidSampleRate(sampleRate string) bool {
	return regexp.MustCompile(`^\d+[smhdwy]$|^\dmo$`).MatchString(sampleRate)
}
