package metrics

const serviceName = "wormscan-api"

type Metrics interface {
	IncExpiredCacheResponse(key string)
	IncOrigin(origin string)
}
