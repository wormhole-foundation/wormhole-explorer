package stats

import "github.com/shopspring/decimal"

type NativeTokenTransferTopAddress struct {
	FromAddress string          `json:"fromAddress"`
	Value       decimal.Decimal `json:"value"`
}
