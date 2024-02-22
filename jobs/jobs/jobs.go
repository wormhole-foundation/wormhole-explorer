// Package jobs define an interface to execute jobs
package jobs

import "context"

// JobIDNotional is the job id for notional job.
const (
	JobIDNotional          = "JOB_NOTIONAL_USD"
	JobIDTransferReport    = "JOB_TRANSFER_REPORT"
	JobIDHistoricalPrices  = "JOB_HISTORICAL_PRICES"
	JobIDMigrationSourceTx = "JOB_MIGRATE_SOURCE_TX"
	JobIDProtocolsStats    = "JOB_PROTOCOLS_STATS"
	JobIDProtocolsActivity = "JOB_PROTOCOLS_ACTIVITY"
)

// Job is the interface for jobs.
type Job interface {
	Run(ctx context.Context) error
}
