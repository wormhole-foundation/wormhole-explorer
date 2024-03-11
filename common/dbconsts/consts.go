package dbconsts

// influx-db constants
const (
	ProtocolsActivityMeasurement = "protocols_activity" // todo:deprecate
	ProtocolsStatsMeasurement    = "protocols_stats_v1" // todo:deprecate

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
