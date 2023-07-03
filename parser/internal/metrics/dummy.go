package metrics

// DummyMetrics is a dummy implementation of Metric interface.
type DummyMetrics struct {
}

// NewDummyMetrics returns a new instance of DummyMetrics.
func NewDummyMetrics() *DummyMetrics {
	return &DummyMetrics{}
}

// IncVaaConsumedQueue increments the number of consumed VAA.
func (d *DummyMetrics) IncVaaConsumedQueue(chainID uint16) {}

// IncVaaUnfiltered increments the number of unfiltered VAA.
func (d *DummyMetrics) IncVaaUnfiltered(chainID uint16) {}

// IncVaaUnexpired increments the number of unexpired VAA.
func (d *DummyMetrics) IncVaaUnexpired(chainID uint16) {}

// IncVaaParsed increments the number of parsed VAA.
func (d *DummyMetrics) IncVaaParsed(chainID uint16) {}

// IncVaaParsedInserted increments the number of parsed VAA inserted into database.
func (d *DummyMetrics) IncVaaParsedInserted(chainID uint16) {}

// IncVaaPayloadParserRequestCount increments the number of vaa payload parser request.
func (d *DummyMetrics) IncVaaPayloadParserRequestCount(chainID uint16) {}

// IncVaaPayloadParserErrorCount increments the number of vaa payload parser error.
func (d *DummyMetrics) IncVaaPayloadParserErrorCount(chainID uint16) {}

// IncVaaPayloadParserSuccessCount increments the number of vaa payload parser success.
func (d *DummyMetrics) IncVaaPayloadParserSuccessCount(chainID uint16) {}

// IncVaaPayloadParserSuccessCount increments the number of vaa payload parser success.
func (d *DummyMetrics) IncVaaPayloadParserNotFoundCount(chainID uint16) {}
