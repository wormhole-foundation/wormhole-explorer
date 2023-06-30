package metrics

const serviceName = "wormscan-parser"

type Metrics interface {
	IncVaaConsumedQueue(chainID uint16)
	IncVaaUnfiltered(chainID uint16)
	IncVaaUnexpired(chainID uint16)
	IncVaaParsed(chainID uint16)
	IncVaaParsedInserted(chainID uint16)

	IncVaaPayloadParserRequestCount(chainID uint16)
	IncVaaPayloadParserErrorCount(chainID uint16)
	IncVaaPayloadParserNotFoundCount(chainID uint16)
	IncVaaPayloadParserSuccessCount(chainID uint16)
}
