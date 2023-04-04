# Backfiller

Reprocess all VAAs in a specified time range:
`STRATEGY_NAME=time_range STRATEGY_TIMESTAMP_AFTER=2023-01-01T00:00:00.000Z STRATEGY_TIMESTAMP_BEFORE=2023-04-01T00:00:00.000Z ./backfiller`

Reprocess only VAAs that failed due to internal errors:
`STRATEGY_NAME=reprocess_failed ./backfiller`