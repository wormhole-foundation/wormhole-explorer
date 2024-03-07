package dbconsts

// influx-db constants
const (
	ProtocolsActivityMeasurement = "protocols_activity"
	ProtocolsStatsMeasurement    = "protocols_stats_v1"

	CctpStatsMeasurementHourly        = intProtocolStatsMeasurement1h
	TokenBridgeStatsMeasurementHourly = intProtocolStatsMeasurement1h
	intProtocolStatsMeasurement1h     = "core_protocols_stats_1h"

	CctpStatsMeasurementDaily        = intProtocolStatsMeasurement1d
	TokenBridgeStatsMeasurementDaily = intProtocolStatsMeasurement1d
	intProtocolStatsMeasurement1d    = "core_protocols_stats_1d"
)
