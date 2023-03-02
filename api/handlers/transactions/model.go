package transactions

import "time"

type Transaction struct {
}

type ChainActivityResult struct {
	Start              time.Time `mapstructure:"_start"`
	End                time.Time `mapstructure:"_stop"`
	ChainSourceID      string    `mapstructure:"chain_source_id"`
	ChainDestinationID string    `mapstructure:"chain_destination_id"`
	Volume             uint64    `mapstructure:"volume"`
}

type ChainActivityQuery struct {
	Start      *time.Time
	End        *time.Time
	AppIDs     []string
	IsNotional bool
}

func (q *ChainActivityQuery) HasAppIDS() bool {
	return len(q.AppIDs) > 0
}

func (q *ChainActivityQuery) GetAppIDs() []string {
	return q.AppIDs
}

func (q *ChainActivityQuery) GetStart() time.Time {
	if q.Start == nil {
		return time.UnixMilli(0)
	}
	return *q.Start
}

func (q *ChainActivityQuery) GetEnd() time.Time {
	if q.End == nil {
		return time.Now()
	}
	return *q.End
}
