package middleware

import (
	"errors"
	"github.com/certusone/wormhole/node/pkg/vaa"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

var (
	ErrMalformedChain    = errors.New("WRONG_CHAIN_ID")
	ErrMalformedAddr     = errors.New("MALFORMED_EMITTER_ADDR")
	ErrMalformedSequence = errors.New("MALFORMED_SEQUENCE_NUMBER")
)

func ExtractChainID(c *fiber.Ctx) (vaa.ChainID, error) {
	chain, err := c.ParamsInt("chain")
	if err != nil {
		return vaa.ChainIDUnset, ErrMalformedChain
	}
	return vaa.ChainID(chain), nil
}

func ExtractEmitterAddr(c *fiber.Ctx) (*vaa.Address, error) {
	emitterStr := c.Params("emitter")
	emitter, err := vaa.StringToAddress(emitterStr)
	if err != nil {
		return nil, ErrMalformedAddr
	}
	return &emitter, nil
}

func ExtractSequence(c *fiber.Ctx) (uint64, error) {
	sequence := c.Params("sequence")
	seq, err := strconv.ParseUint(sequence, 10, 64)
	if err != nil {
		return 0, err
	}
	return seq, nil
}

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
