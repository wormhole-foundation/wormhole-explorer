package metrics

const serviceName = "wormscan-analytics"

type Metrics interface {
	IncFailedMeasurement(measurement string)
	IncSuccessfulMeasurement(measurement string)
	IncMissingNotional(symbol string)
	IncFoundNotional(symbol string)
	IncMissingToken(chain, token string)
	IncFoundToken(chain, token string)
	IncExpiredMessage(chain, source string)
	IncInvalidMessage(chain, source string)
	IncUnprocessedMessage(chain, source string)
	IncProcessedMessage(chain, source string)
}
