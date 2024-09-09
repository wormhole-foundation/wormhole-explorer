package metrics

import (
	"time"
)

const serviceName = "wormscan-parser"

type Metrics interface {
	IncVaaConsumedQueue(chainID uint16)
	IncVaaUnfiltered(chainID uint16)
	IncVaaParsed(chainID uint16)
	IncVaaParsedInserted(chainID uint16)
	IncVaaParsedInsertFailed(chainID uint16, dbLayer string)

	IncVaaPayloadParserRequestCount(chainID uint16)
	IncVaaPayloadParserErrorCount(chainID uint16)
	IncVaaPayloadParserNotFoundCount(chainID uint16)
	IncVaaPayloadParserSuccessCount(chainID uint16)

	IncExpiredMessage(chain, source string)
	IncUnprocessedMessage(chain, source string)
	IncProcessedMessage(chain, source string)

	VaaProcessingDuration(chain string, start *time.Time)
}
