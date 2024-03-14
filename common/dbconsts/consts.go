package dbconsts

// influx-db constants
const (
	ProtocolsActivityMeasurementHourly = "protocols_activity_1h"
	ProtocolsActivityMeasurementDaily  = "protocols_activity_1d"
	ProtocolsStatsMeasurementDaily     = "protocols_stats_1d"
	ProtocolsStatsMeasurementHourly    = "protocols_stats_1h"

	CctpStatsMeasurementHourly        = intProtocolStatsMeasurement1h
	TokenBridgeStatsMeasurementHourly = intProtocolStatsMeasurement1h
	intProtocolStatsMeasurement1h     = "core_protocols_stats_1h"

	CctpStatsMeasurementDaily        = intProtocolStatsMeasurement1d
	TokenBridgeStatsMeasurementDaily = intProtocolStatsMeasurement1d
	intProtocolStatsMeasurement1d    = "core_protocols_stats_1d"
)
