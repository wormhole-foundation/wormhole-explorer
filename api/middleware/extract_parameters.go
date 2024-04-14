// package middleare contains all the middleware function to use in the API.
package middleware

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/stats"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// ExtractChainID get chain parameter from route path.
func ExtractChainID(c *fiber.Ctx, l *zap.Logger) (sdk.ChainID, error) {

	chain, err := c.ParamsInt("chain")
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to get chain parameter",
			zap.Error(err),
			zap.Int("chain", chain),
			zap.String("requestID", requestID),
		)

		return sdk.ChainIDUnset, response.NewInvalidParamError(c, "WRONG CHAIN ID", errors.WithStack(err))
	}

	return sdk.ChainID(chain), nil
}

// ExtractFromChain obtains the "toChain" query parameter from the request.
//
// When the parameter is not present, the function returns: a nil ChainID and a nil error.
func ExtractToChain(c *fiber.Ctx, l *zap.Logger) (*sdk.ChainID, error) {

	param := c.Query("toChain")
	if param == "" {
		return nil, nil
	}

	chain, err := strconv.ParseInt(param, 10, 16)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to parse toChain parameter",
			zap.Error(err),
			zap.String("requestID", requestID),
		)

		return nil, response.NewInvalidParamError(c, "INVALID TO_CHAIN VALUE", errors.WithStack(err))
	}

	result := sdk.ChainID(chain)
	return &result, nil
}

func ExtractChain(c *fiber.Ctx, l *zap.Logger) (*sdk.ChainID, error) {
	return extractChainQueryParam(c, l, "chain")
}

func ExtractSourceChain(c *fiber.Ctx, l *zap.Logger) (*sdk.ChainID, error) {
	return extractChainQueryParam(c, l, "sourceChain")
}

func ExtractTargetChain(c *fiber.Ctx, l *zap.Logger) (*sdk.ChainID, error) {
	return extractChainQueryParam(c, l, "targetChain")
}

func extractChainQueryParam(c *fiber.Ctx, l *zap.Logger, queryParam string) (*sdk.ChainID, error) {

	param := c.Query(queryParam)
	if param == "" {
		return nil, nil
	}

	chain, err := strconv.ParseInt(param, 10, 16)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to parse toChain parameter",
			zap.Error(err),
			zap.String("requestID", requestID),
		)

		return nil, response.NewInvalidParamError(c, "INVALID CHAIN VALUE", errors.WithStack(err))
	}

	result := sdk.ChainID(chain)
	return &result, nil
}

// ExtractEmitterAddr parses the emitter address from the request path.
//
// When the parameter `chainIdHint` is not nil, this function will attempt to parse the
// native address format of the specified chain.
//
// The fallback behavior is to parse the address according to the Wormhole hex format.
func ExtractEmitterAddr(c *fiber.Ctx, l *zap.Logger, chainIdHint *sdk.ChainID) (*types.Address, error) {

	emitterStr := c.Params("emitter")

	// Decide whether to accept the Solana address format based on the context
	var acceptSolanaFormat bool
	if chainIdHint != nil && *chainIdHint == sdk.ChainIDSolana {
		acceptSolanaFormat = true
	}

	// Attempt to parse the address
	emitter, err := types.StringToAddress(emitterStr, acceptSolanaFormat)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to convert emitter to wormhole address",
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

	// Attempt to parse the address
	guardianAddress, err := types.StringToAddress(tmp, false /*acceptSolanaFormat*/)
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
func ExtractVAAChainIDEmitter(c *fiber.Ctx, l *zap.Logger) (sdk.ChainID, *types.Address, error) {

	chainID, err := ExtractChainID(c, l)
	if err != nil {
		return sdk.ChainIDUnset, nil, err
	}

	address, err := ExtractEmitterAddr(c, l, &chainID)
	if err != nil {
		return chainID, nil, err
	}

	return chainID, address, nil
}

// ExtractVAAParams get VAAA chain, address and sequence from route path.
func ExtractVAAParams(c *fiber.Ctx, l *zap.Logger) (sdk.ChainID, *types.Address, uint64, error) {

	chainID, err := ExtractChainID(c, l)
	if err != nil {
		return sdk.ChainIDUnset, nil, 0, err
	}

	address, err := ExtractEmitterAddr(c, l, &chainID)
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
func ExtractObservationSigner(c *fiber.Ctx, l *zap.Logger) (*sdk.Address, error) {

	signer := c.Params("signer")

	signerAddr, err := sdk.StringToAddress(signer)
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

// ExtractAddressFromQueryParams parses the `address` parameter from the query string.
//
// If the parameter is not present, the function returns an empty string
func ExtractAddressFromQueryParams(c *fiber.Ctx, l *zap.Logger) string {
	return c.Query("address")
}

// ExtractAddressFromPath parses the `id` parameter from the route path.
func ExtractAddressFromPath(c *fiber.Ctx, l *zap.Logger) string {
	return c.Params("id")
}

// ExtractQueryParam parses the `q` parameter from query params.
func ExtractQueryParam(c *fiber.Ctx, l *zap.Logger) string {
	return c.Query("q")
}

// GetTxHash parses the `txHash` parameter from query params.
func GetTxHash(c *fiber.Ctx, l *zap.Logger) (*types.TxHash, error) {

	value := c.Query("txHash")
	if value == "" {
		return nil, nil
	}

	txHash, err := types.ParseTxHash(value)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to parse txHash",
			zap.Error(err),
			zap.String("txHash", value),
			zap.String("requestID", requestID),
		)
		return nil, response.NewInvalidParamError(c, "MALFORMED TX HASH", errors.WithStack(err))
	}

	return txHash, nil
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

func ExtractExclusiveAppId(c *fiber.Ctx) (bool, error) {
	query := c.Query("exclusiveAppId")
	if query == "" {
		return false, nil
	}
	return strconv.ParseBool(query)
}

func ExtractTimeSpan(c *fiber.Ctx, l *zap.Logger) (string, error) {
	// get the timeSpan from query params
	timeSpanStr := c.Query("timeSpan", "1d")

	// validate the timeSpan
	if !isValidTimeSpan(timeSpanStr) {
		return "", response.NewInvalidQueryParamError(c, "INVALID <timeSpan> QUERY PARAMETER", nil)
	}
	return timeSpanStr, nil
}

// isValidTimeSpan check that the timeSpan is valid.
func isValidTimeSpan(timeSpan string) bool {
	return regexp.MustCompile(`^1d$|^1w$|^1mo$`).MatchString(timeSpan)
}

func ExtractSampleRate(c *fiber.Ctx, l *zap.Logger) (string, error) {
	// get the sampleRate from query params
	sampleRateStr := c.Query("sampleRate", "1h")

	// validate the sampleRate
	if !isValidSampleRate(sampleRateStr) {
		return "", response.NewInvalidQueryParamError(c, "INVALID <sampleRate> QUERY PARAMETER", nil)
	}
	return sampleRateStr, nil
}

func isValidSampleRate(sampleRate string) bool {
	return regexp.MustCompile(`^1h$|^1d$`).MatchString(sampleRate)
}

func ExtractTimeSpanAndSampleRate(c *fiber.Ctx, l *zap.Logger) (string, string, error) {
	timeSpan, err := ExtractTimeSpan(c, l)
	if err != nil {
		return "", "", err
	}
	sampleRate, err := ExtractSampleRate(c, l)
	if err != nil {
		return "", "", err
	}

	switch timeSpan {
	case "1d":
		if sampleRate != "1h" {
			return "", "", response.NewInvalidQueryParamError(c, "INVALID CONFIGURATION <timeSpan>, <sampleRate> QUERY PARAMETERS.", nil)
		}
	case "1w":
		if sampleRate != "1d" {
			return "", "", response.NewInvalidQueryParamError(c, "INVALID CONFIGURATION <timeSpan>, <sampleRate> QUERY PARAMETERS", nil)
		}
	case "1mo":
		if sampleRate != "1d" {
			return "", "", response.NewInvalidQueryParamError(c, "INVALID CONFIGURATION <timeSpan>, <sampleRate> QUERY PARAMETERS", nil)
		}
	}

	return timeSpan, sampleRate, nil
}

func ExtractTime(c *fiber.Ctx, queryParam string) (*time.Time, error) {
	// get the start_time from query params
	date := c.Query(queryParam, "")
	if date == "" {
		return nil, nil
	}

	t, err := time.Parse("20060102T150405Z", date)
	if err != nil {
		return nil, response.NewInvalidQueryParamError(c, fmt.Sprintf("INVALID <%s> QUERY PARAMETER", queryParam), nil)
	}
	return &t, nil
}

func ExtractApps(ctx *fiber.Ctx) ([]string, error) {
	apps := ctx.Query("apps")
	if apps == "" {
		return nil, nil
	}
	return strings.Split(apps, ","), nil
}

func ExtractIsNotional(ctx *fiber.Ctx) (bool, error) {
	by := ctx.Query("by")
	if by == "" {
		return true, nil
	}
	if by == "notional" {
		return true, nil
	}
	if by == "tx" {
		return false, nil
	}
	return false, response.NewInvalidQueryParamError(ctx, "INVALID <by> QUERY PARAMETER", nil)
}

func ExtractChainActivityTimeSpan(ctx *fiber.Ctx) (transactions.ChainActivityTimeSpan, error) {
	s := ctx.Query("timeSpan", string(transactions.ChainActivityTs7Days))
	timeSpan, err := transactions.ParseChainActivityTimeSpan(s)
	if err != nil {
		return "", response.NewInvalidQueryParamError(ctx, "INVALID <timeSpan> QUERY PARAMETER", nil)
	}
	return timeSpan, nil
}

// ExtractTopStatisticsTimeSpan parses the `timespan` parameter used on top statistics endpoints.
//
// The endpoints that accept this parameter are:
// * `GET /api/v1/top-assets-by-volume`
// * `GET /api/v1/top-chain-pairs-by-num-transfers`
func ExtractTopStatisticsTimeSpan(ctx *fiber.Ctx) (*transactions.TopStatisticsTimeSpan, error) {

	s := ctx.Query("timeSpan")
	timeSpan, err := transactions.ParseTopStatisticsTimeSpan(s)
	if err != nil {
		return nil, response.NewInvalidQueryParamError(ctx, "INVALID <timeSpan> QUERY PARAMETER", nil)
	}

	return timeSpan, nil
}

// ExtractTokenAddress get token address from route path.
func ExtractTokenAddress(c *fiber.Ctx, l *zap.Logger) (*types.Address, error) {
	strTokenAddress := c.Params("token_address")
	tokenAddress, err := types.StringToAddress(strTokenAddress, true)
	if err != nil {
		requestID := fmt.Sprintf("%v", c.Locals("requestid"))
		l.Error("failed to convert string to address",
			zap.Error(err),
			zap.String("token_address", strTokenAddress),
			zap.String("requestID", requestID),
		)
		return nil, response.NewInvalidParamError(c, "MALFORMED TOKEN_ADDRESS", errors.WithStack(err))
	}
	return tokenAddress, nil
}

func ExtractSymbolWithAssetsTimeSpan(ctx *fiber.Ctx) (*stats.SymbolWithAssetsTimeSpan, error) {
	defaultTimeSpan := stats.TimeSpan7Days
	s := ctx.Query("timeSpan")
	if s == "" {
		return &defaultTimeSpan, nil
	}
	timeSpan, err := stats.ParseSymbolsWithAssetsTimeSpan(s)
	if err != nil {
		return nil, response.NewInvalidQueryParamError(ctx, "INVALID <timeSpan> QUERY PARAMETER", nil)
	}

	return timeSpan, nil
}

func ExtractTopCorridorsTimeSpan(ctx *fiber.Ctx) (*stats.TopCorridorsTimeSpan, error) {
	defaultTimeSpan := stats.TimeSpan2DaysTopCorridors
	s := ctx.Query("timeSpan")
	if s == "" {
		return &defaultTimeSpan, nil
	}
	timeSpan, err := stats.ParseTopCorridorsTimeSpan(s)
	if err != nil {
		return nil, response.NewInvalidQueryParamError(ctx, "INVALID <timeSpan> QUERY PARAMETER", nil)
	}

	return timeSpan, nil
}
