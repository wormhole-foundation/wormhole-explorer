package prices

import (
	"math/big"

	"github.com/shopspring/decimal"
)

// CalculatePriceUSD calculates the price in USD for a given notional value and amount of tokens
func CalculatePriceUSD(notionalUSD decimal.Decimal, amount *big.Int, decimals int64) decimal.Decimal {

	var exp int32
	if decimals > 8 {
		exp = 8
	} else {
		exp = int32(decimals)
	}
	tokenAmount := decimal.NewFromBigInt(amount, -exp)

	// Compute the amount in USD
	usdAmount := tokenAmount.Mul(notionalUSD)

	return usdAmount
}
