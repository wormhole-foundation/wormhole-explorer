package metrics

const serviceName = "wormscan-parser"

type Metrics interface {
	IncVaaConsumed(chainID uint16)
	IncVaaUnfiltered(chainID uint16)
	IncVaaUnexpired(chainID uint16)
	IncVaaParsed(chainID uint16)
	IncParsedVaaInserted(chainID uint16)

	IncVaaPayloadParserRequestCount(chainID uint16)
	IncVaaPayloadParserErrorCount(chainID uint16)
	IncVaaPayloadParserSuccessCount(chainID uint16)
}
