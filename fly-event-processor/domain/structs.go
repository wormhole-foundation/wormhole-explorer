package domain

import "time"

type GovernorStatus struct {
	ChainID        uint32
	EmitterAddress string
	Sequence       uint64
	GovernorTxHash string
	ReleaseTime    time.Time
	Amount         uint64
}
