package prices

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Controller struct {
	priceService *PriceService
	logger       *zap.Logger
}

func NewController(priceService *PriceService, logger *zap.Logger) *Controller {
	return &Controller{
		priceService: priceService,
		logger:       logger,
	}
}

func (c *Controller) FindByToken(ctx *fiber.Ctx) error {
	chainID, err := extractChainID(ctx, c.logger)
	if err != nil {
		return err
	}

	tokenAddress, err := extractTokenAddress(ctx, c.logger)
	if err != nil {
		return err
	}

	dateTime, err := extractDatetime(ctx, c.logger)
	if err != nil {
		return err
	}

	price, err := c.priceService.GetPrice(ctx.Context(), *chainID, tokenAddress.Hex(), *dateTime)

	return c.handleResponse(err, ctx, price)
}

func (c *Controller) FindByCoingeckoID(ctx *fiber.Ctx) error {
	coingeckoID, err := extractCoingeckoID(ctx, c.logger)
	if err != nil {
		return err
	}

	dateTime, err := extractDatetime(ctx, c.logger)
	if err != nil {
		return err
	}

	price, err := c.priceService.GetPriceByCoingeckoID(ctx.Context(), coingeckoID, *dateTime)

	if errors.Is(err, ErrTokenNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "token not found")
	}

	return c.handleResponse(err, ctx, price)
}

func (*Controller) handleResponse(err error, ctx *fiber.Ctx, price *Price) error {
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "token not found")
		}
		return err
	}

	return ctx.JSON(price)
}

func extractCoingeckoID(c *fiber.Ctx, l *zap.Logger) (string, error) {
	param := c.Params("coingeckoId")
	if param == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, "missing coingeckoId parameter")
	}
	return param, nil
}

func extractChainID(c *fiber.Ctx, l *zap.Logger) (*sdk.ChainID, error) {
	param := c.Params("tokenChainId")
	if param == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "missing tokenChainId parameter")
	}

	chain, err := strconv.ParseInt(param, 10, 16)
	if err != nil {
		l.Error("failed to parse tokenChainId parameter",
			zap.Error(err),
		)
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid chain id parameter")
	}
	result := sdk.ChainID(chain)
	return &result, nil
}

func extractTokenAddress(c *fiber.Ctx, l *zap.Logger) (*types.Address, error) {
	tokenAddressStr := c.Params("tokenAddress")
	tokenAddress, err := types.StringToAddress(tokenAddressStr, true)
	if err != nil {
		l.Error("failed to convert token address to wormhole address",
			zap.Error(err),
			zap.String("tokenAddressStr", tokenAddressStr))
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid token address")
	}

	return tokenAddress, nil
}

func extractDatetime(c *fiber.Ctx, l *zap.Logger) (*time.Time, error) {
	datetime := c.Params("datetime")
	if datetime == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, "missing datetime parameter")
	}
	v, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid datetime parameter")
	}
	if v.After(time.Now()) {
		return nil, fiber.NewError(fiber.StatusBadRequest, "datetime must be in the past")
	}
	return &v, nil
}
