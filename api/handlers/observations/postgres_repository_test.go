package observations

import (
	"testing"

	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"gotest.tools/assert"
)

func TestPostgresRepository_toQueryEmpty(t *testing.T) {

	q := Query()

	query, params := q.toQuery()
	assert.Equal(t, 0, len(params))
	assert.Equal(t,
		`SELECT obs.* FROM wormholescan.wh_observations obs 
WHERE 1 = 1 
ORDER BY obs.created_at DESC 
LIMIT 50 OFFSET 0`, query)
}

func TestPostgresRepository_toQueryWithChain(t *testing.T) {

	q := Query().SetChain(sdk.ChainIDSolana)

	query, params := q.toQuery()
	assert.Equal(t, 1, len(params))
	assert.Equal(t,
		`SELECT obs.* FROM wormholescan.wh_observations obs 
WHERE obs.emitter_chain_id = $1 
ORDER BY obs.created_at DESC 
LIMIT 50 OFFSET 0`, query)
}

func TestPostgresRepository_toQueryWithTxHash(t *testing.T) {
	txHash, _ := types.ParseTxHash("efb32ca3f7b529d00c18e928b45189be5b534d1ae5f35697a1dce6005713cf0b")
	q := Query().SetTxHash(txHash)
	query, params := q.toQuery()
	assert.Equal(t, 1, len(params))
	assert.Equal(t,
		`SELECT obs.* FROM wormholescan.wh_observations obs 
JOIN wormholescan.wh_operation_transactions ot ON obs.hash = ot.attestation_id 
WHERE ot.tx_hash = $1 
ORDER BY obs.created_at DESC 
LIMIT 50 OFFSET 0`, query)
}

func TestPostgresRepository_toQueryWithEmitterAddressAndSequence(t *testing.T) {
	q := Query().SetEmitter("3c1dbf954f2dc8810f1446319025ac1a5ea89328191414f0c2571b6e24e8855c").SetSequence("500")
	query, params := q.toQuery()
	assert.Equal(t, 2, len(params))
	assert.Equal(t,
		`SELECT obs.* FROM wormholescan.wh_observations obs 
WHERE obs.emitter_address = $1 AND obs.sequence = $2 
ORDER BY obs.created_at DESC 
LIMIT 50 OFFSET 0`, query)
}
