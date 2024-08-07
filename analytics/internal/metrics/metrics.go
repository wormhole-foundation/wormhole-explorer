package metrics

import "time"

const serviceName = "wormscan-analytics"

type Metrics interface {
	IncFailedMeasurement(measurement string)
	IncSuccessfulMeasurement(measurement string)
	IncMissingNotional(symbol string)
	IncFoundNotional(symbol string)
	IncMissingToken(chain, token string)
	IncFoundToken(chain, token string)
	IncExpiredMessage(chain, source string, retry uint8)
	IncInvalidMessage(chain, source string, retry uint8)
	IncUnprocessedMessage(chain, source string, retry uint8)
	IncProcessedMessage(chain, source string, retry uint8)
	VaaProcessingDuration(chain string, start *time.Time)
}
