package transactions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type totalPythResult struct {
	ID       string `data:"id"`
	Sequence uint64 `data:"sequence"`
}

type PostgresRepository struct {
	p2pNetwork string
	db         *db.DB
	logger     *zap.Logger
}

func NewPostgresRepository(p2pNetwork string, db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		p2pNetwork: p2pNetwork,
		db:         db,
		logger:     logger.With(zap.String("module", "PostgresTransactionsRepository")),
	}
}

// getTotalPythMessage returns the last sequence for the pyth emitter address
func (r *PostgresRepository) getTotalPythMessage(ctx context.Context) (string, error) {
	if r.p2pNetwork != config.P2pMainNet {
		return "0", nil

	}
	query := `
	SELECT id, sequence FROM wormholescan.wh_attestation_vaas_pythnet 
    WHERE emitter_address = $1
    ORDER BY "timestamp" DESC 
    LIMIT 1`

	var result totalPythResult
	err := r.db.SelectOne(ctx, &result, query, pythEmitterAddr)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return "", err
	}

	return strconv.FormatUint(result.Sequence, 10), nil
}

type operationTxResult struct {
	ChainID           uint16         `db:"chain_id"`
	TxHash            string         `db:"tx_hash"`
	Type              string         `db:"type"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
	AttestationVaasID string         `db:"attestation_vaas_id"`
	MessageID         string         `db:"message_id"`
	Status            *string        `db:"status"`
	FromAddress       *string        `db:"from_address"`
	ToAddress         *string        `db:"to_address"`
	BlockNumber       *string        `db:"block_number"`
	BlockchainMethod  *string        `db:"blockchain_method"`
	FeeDetail         map[string]any `db:"fee_detail"`
	Timestamp         time.Time      `db:"timestamp"`
	RPCResponse       map[string]any `db:"rpc_response"`
}

// FindGlobalTransactionByID returns a global transaction by its ID.
func (r *PostgresRepository) FindGlobalTransactionByID(
	ctx context.Context,
	q *GlobalTransactionQuery,
) (*GlobalTransactionDoc, error) {

	query := `SELECT chain_id, tx_hash, "type", created_at, updated_at, attestation_vaas_id, message_id, status, 
	from_address, to_address, block_number, blockchain_method, fee_detail, timestamp, rpc_response 
	FROM wormholescan.wh_operation_transactions 
	WHERE message_id = $1
	`

	var operationTxs []*operationTxResult
	err := r.db.Select(ctx, &operationTxs, query, q.id)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return nil, err
	}

	globalTransaction, err := createGlobalTransaction(operationTxs)
	if err != nil {
		r.logger.Error("failed to create global transaction", zap.Error(err))
		return nil, err
	}

	return globalTransaction, nil
}

func createGlobalTransaction(operationTxs []*operationTxResult) (*GlobalTransactionDoc, error) {
	if len(operationTxs) == 0 {
		return nil, errors.ErrNotFound
	}

	// check all the operation txs have the same message id
	messageID := operationTxs[0].MessageID
	for _, tx := range operationTxs {
		if tx.MessageID != messageID {
			return nil, fmt.Errorf("operation txs have different message ids %s and %s", messageID, tx.MessageID)
		}
	}

	// create origin and destination txs
	originTx := createOriginTx(operationTxs)
	destinationTx := createDestinationTx(operationTxs)

	// create global transaction
	return &GlobalTransactionDoc{
		ID:            messageID,
		OriginTx:      originTx,
		DestinationTx: destinationTx,
	}, nil
}

func createOriginTx(operationTxs []*operationTxResult) *OriginTx {
	var sourceTx *operationTxResult
	var nestedSourceTx *operationTxResult
	for _, tx := range operationTxs {
		if tx.Type == "source-tx" {
			sourceTx = tx
		}
		if tx.Type == "nested-source-tx" {
			nestedSourceTx = tx
		}
	}

	// denormalize tx hash for compatibility reasons.
	denormalizedTxHash := domain.DenormalizeTxHashByChainId(
		sdk.ChainID(sourceTx.ChainID), sourceTx.TxHash)

	// denormalize from address for compatibility reasons.
	var denormalizedFromAddress string
	if sourceTx.FromAddress != nil {
		denormalizedFromAddress = domain.DenormalizeAddressByChainId(
			sdk.ChainID(sourceTx.ChainID), *sourceTx.FromAddress)
	}

	originTx := &OriginTx{
		TxHash: denormalizedTxHash,
		From:   denormalizedFromAddress,
		Status: *sourceTx.Status,
	}

	var attribute *AttributeDoc
	if nestedSourceTx != nil {

		var denormalizedNestedTxHash string
		if nestedSourceTx.TxHash != "" {
			// denormalize tx hash for compatibility reasons.
			denormalizedNestedTxHash = domain.DenormalizeTxHashByChainId(
				sdk.ChainID(nestedSourceTx.ChainID), nestedSourceTx.TxHash)
		}
		values := map[string]any{
			"originChainId": nestedSourceTx.ChainID,
			"originTxHash":  denormalizedNestedTxHash,
		}
		if nestedSourceTx.FromAddress != nil {
			// denormalize from address for compatibility reasons.
			denormalizedOriginAddress := domain.DenormalizeAddressByChainId(
				sdk.ChainID(nestedSourceTx.ChainID), *nestedSourceTx.FromAddress)
			values["originAddress"] = denormalizedOriginAddress
		}

		attribute = &AttributeDoc{
			Type:  "wormchain-gateway",
			Value: values,
		}
	}
	originTx.Attribute = attribute
	return originTx
}

func createDestinationTx(operationTxs []*operationTxResult) *DestinationTx {
	var destinationTx *operationTxResult
	for _, tx := range operationTxs {
		if tx.Type == "target" {
			destinationTx = tx
		}
	}

	if destinationTx == nil {
		return nil
	}

	var blockNumber string
	if destinationTx.BlockNumber != nil {
		blockNumber = *destinationTx.BlockNumber
	}

	var timestamp *time.Time
	if !destinationTx.Timestamp.IsZero() {
		timestamp = &destinationTx.Timestamp
	}

	var updatedAt *time.Time
	if !destinationTx.UpdatedAt.IsZero() {
		updatedAt = &destinationTx.UpdatedAt
	}

	// denormalize tx hash for compatibility reasons.
	denormalizedTxHash := domain.DenormalizeTxHashByChainId(
		sdk.ChainID(destinationTx.ChainID), destinationTx.TxHash)

	// denormalize from address for compatibility reasons.
	var denormalizedFromAddress string
	if destinationTx.FromAddress != nil {
		denormalizedFromAddress = domain.DenormalizeAddressByChainId(
			sdk.ChainID(destinationTx.ChainID), *destinationTx.FromAddress)
	}

	// denormalize to address for compatibility reasons.
	var denormalizedToAddress string
	if destinationTx.ToAddress != nil {
		denormalizedToAddress = domain.DenormalizeAddressByChainId(
			sdk.ChainID(destinationTx.ChainID), *destinationTx.ToAddress)
	}

	return &DestinationTx{
		ChainID:     sdk.ChainID(destinationTx.ChainID),
		Status:      *destinationTx.Status,
		Method:      *destinationTx.BlockchainMethod,
		TxHash:      denormalizedTxHash,
		From:        denormalizedFromAddress,
		To:          denormalizedToAddress,
		BlockNumber: blockNumber,
		Timestamp:   timestamp,
		UpdatedAt:   updatedAt,
	}
}

type transactionResult struct {
	TransactionChainID          uint16           `db:"transaction_chain_id"`
	TransactionTxHash           string           `db:"transaction_tx_hash"`
	TransactionType             string           `db:"transaction_type"`
	TransactionStatus           *string          `db:"transaction_status"`
	TransactionFromAddress      *string          `db:"transaction_from_address"`
	TransactionToAddress        *string          `db:"transaction_to_address"`
	TransactionTimestamp        time.Time        `db:"transaction_timestamp"`
	TransactionBlockchainMethod *string          `db:"transaction_blockchain_method"`
	TransactionBlockNumber      *string          `db:"transaction_block_number"`
	TransactionFeeDetail        *json.RawMessage `db:"transaction_fee_detail"`

	PriceSymbol     *string          `db:"price_symbol"`
	PriceTotalUSD   *decimal.Decimal `db:"price_total_usd"`
	PriceTotalToken *decimal.Decimal `db:"price_total_token"`

	VaaId               string     `db:"vaa_vaa_id"`
	VaaVersion          uint8      `db:"vaa_version"`
	VaaEmitterChainID   uint16     `db:"vaa_emitter_chain_id"`
	VaaEmitterAddress   string     `db:"vaa_emitter_address"`
	VaaSequence         string     `db:"vaa_sequence"`
	VaaGuardianSetIndex *uint32    `db:"vaa_guardian_set_index"`
	VaaRaw              []byte     `db:"vaa_raw"`
	VaaTimestamp        time.Time  `db:"vaa_timestamp"`
	VaaUpdatedAt        *time.Time `db:"vaa_updated_at"`
	VaaCreatedAt        time.Time  `db:"vaa_created_at"`
	VaaIsDuplicated     bool       `db:"vaa_is_duplicated"`

	PropertiesPayload           *json.RawMessage `db:"properties_payload"`
	PropertiesRawStandardFields *json.RawMessage `db:"properties_raw_standard_fields"`
}

// FindTransactions returns transactions matching a specified search criteria.
func (r *PostgresRepository) FindTransactions(
	ctx context.Context,
	input *FindTransactionsInput,
) ([]TransactionDto, error) {

	if input == nil {
		return nil, errors.ErrInternalError
	}

	// find transaction by id
	if input.id != "" {
		// find transaction by id
		return r.findTransactionById(ctx, input)
	}

	return nil, errors.ErrNotFound
}

func (r *PostgresRepository) findTransactionById(ctx context.Context, input *FindTransactionsInput) ([]TransactionDto, error) {
	query := `SELECT 	
	wot.chain_id as transaction_chain_id,
	wot.tx_hash as transaction_tx_hash,
	wot."type" as transaction_type,
	wot.status as transaction_status,
	wot.from_address as transaction_from_address,
	wot.to_address as transaction_to_address,
	wot.timestamp as transaction_timestamp,
	wot.blockchain_method as transaction_blockchain_method,
	wot.block_number as transaction_block_number,
	wot.fee_detail as transaction_fee_detail,
	wop.symbol as price_symbol,
	wop.total_usd as price_total_usd,
	wop.total_token as price_total_token,
	wav.vaa_id as vaa_vaa_id,
	wav.version as vaa_version,
	wav.emitter_chain_id as vaa_emitter_chain_id,
	wav.emitter_address as vaa_emitter_address,
	wav.sequence as vaa_sequence,
	wav.guardian_set_index as vaa_guardian_set_index,
	wav.raw as vaa_raw,
	wav.timestamp as vaa_timestamp,
	wav.updated_at as vaa_updated_at,
	wav.created_at as vaa_created_at,
	wav.is_duplicated as vaa_is_duplicated,
	wavp.payload as properties_payload,
	wavp.raw_standard_fields as properties_raw_standard_fields
FROM wormholescan.wh_attestation_vaas wav
LEFT JOIN wormholescan.wh_operation_transactions wot ON  wav.id = wot.attestation_vaas_id
LEFT JOIN wormholescan.wh_operation_prices wop ON wop.id = wot.attestation_vaas_id 
LEFT JOIN wormholescan.wh_attestation_vaa_properties wavp ON wavp.id = wot.attestation_vaas_id 
WHERE wav.vaa_id = $1 AND  wav.active = true`

	var txs []*transactionResult
	err := r.db.Select(ctx, &txs, query, input.id)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err),
			zap.String("query", query),
			zap.String("vaa_id", input.id))
		return nil, err
	}

	if len(txs) == 0 {
		return nil, errors.ErrNotFound
	}

	transactionDto, err := r.toTransactionDto(txs)
	if err != nil {
		return nil, err
	}

	return []TransactionDto{*transactionDto}, nil
}

func (r *PostgresRepository) toTransactionDto(txs []*transactionResult) (*TransactionDto, error) {
	if len(txs) == 0 {
		return nil, errors.ErrNotFound
	}

	// check all the transactionResult have the same vaaId
	vaaId := txs[0].VaaId
	for _, t := range txs {
		if t.VaaId != vaaId {
			return nil, fmt.Errorf("transactionResults have different vaa ids %s and %s", vaaId, t.VaaId)
		}
	}

	emitterChain := sdk.ChainID(txs[0].VaaEmitterChainID)
	emitterAddr := txs[0].VaaEmitterAddress
	timestamp := txs[0].VaaTimestamp

	var sourceTx *OriginTx
	var nestedSourceTx *AttributeDoc
	var destTx *DestinationTx
	var price *transactionPrices
	var payload map[string]any
	var properties map[string]any

	for _, t := range txs {

		// get source tx
		if sourceTx == nil {
			sourceTxDto, err := r.toOriginTx(t)
			if err != nil {
				return nil, err
			}
			if sourceTxDto != nil {
				sourceTx = sourceTxDto
			}
		}

		// get nested source tx
		if nestedSourceTx == nil {
			nestedSourceTxDto, err := r.toNestedSourceTx(t)
			if err != nil {
				return nil, err
			}
			if nestedSourceTxDto != nil {
				nestedSourceTx = nestedSourceTxDto
			}
		}

		// get destination tx
		if destTx == nil {
			destTxDto, err := r.toDestinationTx(t)
			if err != nil {
				return nil, err
			}
			if destTxDto != nil {
				destTx = destTxDto
			}
		}

		// get price
		if price == nil {
			priceDto := r.toPrices(t)
			if priceDto != nil {
				price = priceDto
			}
		}

		// get payload
		if payload == nil {
			payloadDto, err := r.toPayload(t)
			if err != nil {
				return nil, err
			}
			if payloadDto != nil {
				payload = *payloadDto
			}
		}

		// get properties
		if properties == nil {
			propertiesDto, err := r.toStandardizedProperties(t)
			if err != nil {
				return nil, err
			}
			if propertiesDto != nil {
				properties = *propertiesDto
			}
		}
	}

	var txHash string
	if sourceTx != nil {
		txHash = sourceTx.TxHash
	}
	if nestedSourceTx != nil && sourceTx != nil {
		sourceTx.Attribute = nestedSourceTx
	}

	globalTransaction := &GlobalTransactionDoc{
		ID:            vaaId,
		OriginTx:      sourceTx,
		DestinationTx: destTx,
	}
	globalTransactions := []GlobalTransactionDoc{*globalTransaction}

	return &TransactionDto{
		ID:                     vaaId,
		EmitterChain:           emitterChain,
		EmitterAddr:            emitterAddr,
		TxHash:                 txHash,
		Timestamp:              timestamp,
		Symbol:                 price.Symbol,
		UsdAmount:              price.TotalUSD,
		TokenAmount:            price.TotalToken,
		GlobalTransations:      globalTransactions,
		Payload:                payload,
		StandardizedProperties: properties,
	}, nil
}

type transactionPrices struct {
	Symbol     string
	TotalUSD   string
	TotalToken string
}

func (r *PostgresRepository) toPrices(t *transactionResult) *transactionPrices {

	var priceSymbol, priceUsdAmount, priceTokenAmount string

	if t.PriceSymbol != nil {
		priceSymbol = *t.PriceSymbol
	}
	if t.PriceTotalUSD != nil {
		priceUsdAmount = t.PriceTotalUSD.String()
	}
	if t.PriceTotalToken != nil {
		priceTokenAmount = t.PriceTotalToken.String()
	}

	return &transactionPrices{
		Symbol:     priceSymbol,
		TotalUSD:   priceUsdAmount,
		TotalToken: priceTokenAmount,
	}
}

func (r *PostgresRepository) toPayload(t *transactionResult) (*map[string]any, error) {

	var payload map[string]any
	if t.PropertiesPayload == nil {
		return nil, nil
	}
	err := json.Unmarshal(*t.PropertiesPayload, &payload)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

func (r *PostgresRepository) toStandardizedProperties(t *transactionResult) (*map[string]any, error) {
	var properties map[string]any
	if t.PropertiesRawStandardFields == nil {
		return nil, nil
	}
	err := json.Unmarshal(*t.PropertiesRawStandardFields, &properties)
	if err != nil {
		return nil, err
	}
	return &properties, nil
}

func (r *PostgresRepository) toOriginTx(t *transactionResult) (*OriginTx, error) {
	if t.TransactionType == "source-tx" {
		chainID := sdk.ChainID(t.TransactionChainID)
		var from string
		if t.TransactionFromAddress != nil {
			from = domain.DenormalizeAddressByChainId(chainID, *t.TransactionFromAddress)
		}
		var status string
		if t.TransactionStatus != nil {
			status = *t.TransactionStatus
		}
		denormalizedTxHash := domain.DenormalizeTxHashByChainId(chainID, t.TransactionTxHash)
		return &OriginTx{
			TxHash: denormalizedTxHash,
			From:   from,
			Status: status,
		}, nil
	}
	return nil, nil
}

func (r *PostgresRepository) toDestinationTx(t *transactionResult) (*DestinationTx, error) {
	if t.TransactionType == "target-tx" {
		chainID := sdk.ChainID(t.TransactionChainID)
		if chainID.String() == sdk.ChainIDUnset.String() {
			return nil, fmt.Errorf("invalid chain id %d for destination tx", t.TransactionChainID)
		}
		var status string
		if t.TransactionStatus != nil {
			status = *t.TransactionStatus
		}
		var method string
		if t.TransactionBlockchainMethod != nil {
			method = *t.TransactionBlockchainMethod
		}
		var from string
		if t.TransactionFromAddress != nil {
			from = domain.DenormalizeAddressByChainId(chainID, *t.TransactionFromAddress)
		}
		var to string
		if t.TransactionToAddress != nil {
			to = domain.DenormalizeAddressByChainId(chainID, *t.TransactionToAddress)
		}
		var blockNumber string
		if t.TransactionBlockNumber != nil {
			blockNumber = *t.TransactionBlockNumber
		}

		return &DestinationTx{
			ChainID:     chainID,
			Status:      status,
			Method:      method,
			TxHash:      domain.DenormalizeTxHashByChainId(chainID, t.TransactionTxHash),
			From:        from,
			To:          to,
			BlockNumber: blockNumber,
			Timestamp:   &t.TransactionTimestamp,
		}, nil
	}
	return nil, nil
}

func (r *PostgresRepository) toNestedSourceTx(t *transactionResult) (*AttributeDoc, error) {

	var attribute *AttributeDoc
	if t.TransactionType == "nested-source-tx" {
		var denormalizedNestedTxHash string
		chainID := sdk.ChainID(t.TransactionChainID)
		if t.TransactionTxHash != "" {
			// denormalize tx hash for compatibility reasons.
			denormalizedNestedTxHash = domain.DenormalizeTxHashByChainId(chainID, t.TransactionTxHash)
		}
		values := map[string]any{
			"originChainId": t.TransactionChainID,
			"originTxHash":  denormalizedNestedTxHash,
		}
		if t.TransactionFromAddress != nil && *t.TransactionFromAddress != "" {
			// denormalize from address for compatibility reasons.
			denormalizedOriginAddress := domain.DenormalizeAddressByChainId(
				chainID, *t.TransactionFromAddress)
			values["originAddress"] = denormalizedOriginAddress
		}

		attribute = &AttributeDoc{
			Type:  "wormchain-gateway",
			Value: values,
		}
		return attribute, nil
	}
	return nil, nil
}
