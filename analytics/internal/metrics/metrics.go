package metrics

const serviceName = "wormscan-analytics-v2"

type Metrics interface {
	IncFailedMeasurement(measurement string)
	IncSuccessfulMeasurement(measurement string)
	IncMissingNotional(symbol string)
	IncFoundNotional(symbol string)
	IncMissingToken(chain, token string)
	IncFoundToken(chain, token string)
}
