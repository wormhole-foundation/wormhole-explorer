package metrics

import "time"

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

// IncVaaAttestationPropertiesInserted increment the number of attestation properties inserted into database.
func (d *DummyMetrics) IncVaaAttestationPropertiesInserted(chainID uint16) {}

// IncParseVaaInserted increments the number of parsed VAA inserted into database.
func (d *DummyMetrics) IncParseVaaInserted(chainID uint16) {}

// IncVaaPayloadParserRequestCount increments the number of vaa payload parser request.
func (d *DummyMetrics) IncVaaPayloadParserRequestCount(chainID uint16) {}

// IncVaaPayloadParserErrorCount increments the number of vaa payload parser error.
func (d *DummyMetrics) IncVaaPayloadParserErrorCount(chainID uint16) {}

// IncVaaPayloadParserSuccessCount increments the number of vaa payload parser success.
func (d *DummyMetrics) IncVaaPayloadParserSuccessCount(chainID uint16) {}

// IncVaaPayloadParserSuccessCount increments the number of vaa payload parser success.
func (d *DummyMetrics) IncVaaPayloadParserNotFoundCount(chainID uint16) {}

// IncExpiredMessage increments the number of expired message.
func (p *DummyMetrics) IncExpiredMessage(chain, source string) {}

// IncUnprocessedMessage increments the number of unprocessed message.
func (p *DummyMetrics) IncUnprocessedMessage(chain, source string) {}

// IncProcessedMessage increments the number of processed message.
func (p *DummyMetrics) IncProcessedMessage(chain, source string) {}

// VaaProcessingDuration increments the duration of VAA processing.
func (m *DummyMetrics) VaaProcessingDuration(chain string, start *time.Time) {}
