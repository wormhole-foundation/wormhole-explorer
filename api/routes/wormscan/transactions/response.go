package transactions

type Tx struct {
	Chain        int           `json:"chain"`
	Volume       uint64        `json:"volume"`
	Percentage   float64       `json:"percentage"`
	Destinations []Destination `json:"destinations"`
}

type Destination struct {
	Chain      int     `json:"chain"`
	Volume     uint64  `json:"volume"`
	Percentage float64 `json:"percentage"`
}

// ChainActivity represent a cross chain activity.
type ChainActivity struct {
	Txs []Tx `json:"txs"`
}
