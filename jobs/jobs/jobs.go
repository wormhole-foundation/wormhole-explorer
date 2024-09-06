// Package jobs define an interface to execute jobs
package jobs

import "context"

// JobIDNotional is the job id for notional job.
const (
	JobIDNotional              = "JOB_NOTIONAL_USD"
	JobIDTransferReport        = "JOB_TRANSFER_REPORT"
	JobIDHistoricalPrices      = "JOB_HISTORICAL_PRICES"
	JobIDMigrationSourceTx     = "JOB_MIGRATE_SOURCE_TX"
	JobIDProtocolsStatsDaily   = "JOB_PROTOCOLS_STATS_DAILY"
	JobIDProtocolsStatsHourly  = "JOB_PROTOCOLS_STATS_HOURLY"
	JobIDNTTTopAddressStats    = "JOB_NTT_TOP_ADDRESS_STATS"
	JobIDNTTTopHolderStats     = "JOB_NTT_TOP_HOLDER_STATS"
	JobIDNTTMedianStats        = "JOB_NTT_MEDIAN_STATS"
	JobIDMigrationNativeTxHash = "JOB_MIGRATE_NATIVE_TX_HASH"
)

// Job is the interface for jobs.
type Job interface {
	Run(ctx context.Context) error
}
