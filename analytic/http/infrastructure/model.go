package infrastructure

// InfluxStatus represent a influx server status.
type InfluxStatus struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}
