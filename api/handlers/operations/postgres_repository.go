package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type operationResult struct {
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

	VaaVersion          *uint8     `db:"vaa_version"`
	VaaEmitterChainID   *uint16    `db:"vaa_emitter_chain_id"`
	VaaEmitterAddress   *string    `db:"vaa_emitter_address"`
	VaaSequence         *string    `db:"vaa_sequence"`
	VaaGuardianSetIndex *uint32    `db:"vaa_guardian_set_index"`
	VaaRaw              []byte     `db:"vaa_raw"`
	VaaTimestamp        *time.Time `db:"vaa_timestamp"`
	VaaUpdatedAt        *time.Time `db:"vaa_updated_at"`
	VaaCreatedAt        *time.Time `db:"vaa_created_at"`
	VaaIsDuplicated     *bool      `db:"vaa_is_duplicated"`

	PropertiesPayload           *json.RawMessage `db:"properties_payload"`
	PropertiesRawStandardFields *json.RawMessage `db:"properties_raw_standard_fields"`
}

type PostgresRepository struct {
	db     *db.DB
	logger *zap.Logger
}

func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{db: db,
		logger: logger.With(zap.String("module", "PostgresOperationRepository"))}
}

// FindById returns the operations for the given chainID/emitter/seq.
func (r *PostgresRepository) FindById(ctx context.Context, messageID string) (*OperationDto, error) {
	query := `
SELECT 	
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
FROM wormholescan.wh_operation_transactions wot 
LEFT JOIN wormholescan.wh_attestation_vaas wav ON  wav.id = wot.attestation_vaas_id AND wav.active = true
LEFT JOIN wormholescan.wh_operation_prices wop ON wop.id = wot.attestation_vaas_id 
LEFT JOIN wormholescan.wh_attestation_vaa_properties wavp ON wavp.id = wot.attestation_vaas_id 
WHERE wot.message_id = $1
	`
	var ops []*operationResult
	err := r.db.Select(ctx, &ops, query, messageID)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query), zap.String("message_id", messageID))
		return nil, err
	}

	if len(ops) == 0 {
		return nil, errors.ErrNotFound
	}

	return r.toOperationDto(ops, messageID)
}

func (r *PostgresRepository) FindAll(ctx context.Context, filter OperationFilter) ([]*OperationDto, error) {
	return nil, nil
}

func (r *PostgresRepository) toOperationDto(ops []*operationResult, messageID string) (*OperationDto, error) {

	var sourceTx *OriginTx
	var nestedSourceTx *AttributeDoc
	var destTx *DestinationTx
	var vaa *VaaDto
	var payload *map[string]any
	var standardizedProperties *StandardizedProperties
	var price *operationPrices

	for _, op := range ops {

		if sourceTx == nil {
			sourceTxDto, err := r.toOriginTx(op)
			if err != nil {
				return nil, err
			}
			if sourceTxDto != nil {
				sourceTx = sourceTxDto
			}
		}

		if nestedSourceTx == nil {
			nestedSourceTxDto, err := r.toNestedSourceTx(op)
			if err != nil {
				return nil, err
			}
			if nestedSourceTxDto != nil {
				nestedSourceTx = nestedSourceTxDto
			}
		}

		if destTx == nil {
			destTxDto, err := r.toDestinationTx(op)
			if err != nil {
				return nil, err
			}
			if destTxDto != nil {
				destTx = destTxDto
			}
		}

		if price == nil {
			priceDto := r.toPrices(op)
			if priceDto != nil {
				price = priceDto
			}
		}

		if vaa == nil {
			vaaDto := r.toVaaDto(op, messageID)
			if vaaDto != nil {
				vaa = vaaDto
			}
		}

		if payload == nil {
			payloadDto, err := r.toPayload(op)
			if err != nil {
				return nil, err
			}
			if payloadDto != nil {
				payload = payloadDto
			}
		}

		if standardizedProperties == nil {
			standardizedPropertiesDto, err := r.toStandardizedProperties(op)
			if err != nil {
				return nil, err
			}
			if standardizedPropertiesDto != nil {
				standardizedProperties = standardizedPropertiesDto
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
	result := &OperationDto{
		ID:                     messageID,
		TxHash:                 txHash,
		Symbol:                 price.Symbol,
		UsdAmount:              price.TotalUSD,
		TokenAmount:            price.TotalToken,
		Vaa:                    vaa,
		SourceTx:               sourceTx,
		DestinationTx:          destTx,
		Payload:                *payload,
		StandardizedProperties: standardizedProperties,
	}
	return result, nil
}

func (r *PostgresRepository) toOriginTx(op *operationResult) (*OriginTx, error) {
	if op.TransactionType == "source-tx" {
		chainID := sdk.ChainID(op.TransactionChainID)
		var from string
		if op.TransactionFromAddress != nil {
			from = domain.DenormalizeAddressByChainId(chainID, *op.TransactionFromAddress)
		}
		var status string
		if op.TransactionStatus != nil {
			status = *op.TransactionStatus
		}
		//Attribute *AttributeDoc `bson:"attribute" json:"attribute"`
		var fee *FeeDoc
		if op.TransactionFeeDetail != nil {
			err := json.Unmarshal(*op.TransactionFeeDetail, &fee)
			if err != nil {
				return nil, err
			}
		}

		denormalizedTxHash := domain.DenormalizeTxHashByChainId(chainID, op.TransactionTxHash)
		return &OriginTx{
			TxHash:    denormalizedTxHash,
			From:      from,
			Status:    status,
			Timestamp: &op.TransactionTimestamp,
			Fee:       fee,
		}, nil
	}
	return nil, nil
}

func (r *PostgresRepository) toDestinationTx(op *operationResult) (*DestinationTx, error) {
	if op.TransactionType == "target-tx" {
		chainID := sdk.ChainID(op.TransactionChainID)
		if chainID.String() == sdk.ChainIDUnset.String() {
			return nil, fmt.Errorf("invalid chain id %d for destination tx", op.TransactionChainID)
		}
		var status string
		if op.TransactionStatus != nil {
			status = *op.TransactionStatus
		}
		var method string
		if op.TransactionBlockchainMethod != nil {
			method = *op.TransactionBlockchainMethod
		}
		var from string
		if op.TransactionFromAddress != nil {
			from = domain.DenormalizeAddressByChainId(chainID, *op.TransactionFromAddress)
		}
		var to string
		if op.TransactionToAddress != nil {
			to = domain.DenormalizeAddressByChainId(chainID, *op.TransactionToAddress)
		}
		var blockNumber string
		if op.TransactionBlockNumber != nil {
			blockNumber = *op.TransactionBlockNumber
		}

		//Attribute *AttributeDoc `bson:"attribute" json:"attribute"`
		var fee *FeeDoc
		if op.TransactionFeeDetail != nil {
			err := json.Unmarshal(*op.TransactionFeeDetail, &fee)
			if err != nil {
				return nil, err
			}
		}

		return &DestinationTx{
			ChainID:     chainID,
			Status:      status,
			Method:      method,
			TxHash:      domain.DenormalizeTxHashByChainId(chainID, op.TransactionTxHash),
			From:        from,
			To:          to,
			BlockNumber: blockNumber,
			Timestamp:   &op.TransactionTimestamp,
			Fee:         fee,
		}, nil
	}
	return nil, nil
}

func (r *PostgresRepository) toVaaDto(op *operationResult, messageID string) *VaaDto {

	if op.VaaVersion == nil {
		return nil
	}
	version := *op.VaaVersion

	if op.VaaEmitterChainID == nil {
		return nil
	}
	emitterChainID := sdk.ChainID(*op.VaaEmitterChainID)

	if op.VaaEmitterAddress == nil {
		return nil
	}
	emitterAddr := *op.VaaEmitterAddress

	if op.VaaSequence == nil {
		return nil
	}
	sequence := *op.VaaSequence

	if op.VaaGuardianSetIndex == nil {
		return nil
	}
	guardianSetIndex := *op.VaaGuardianSetIndex

	if op.VaaRaw == nil {
		return nil
	}
	raw := op.VaaRaw

	if op.VaaTimestamp == nil {
		return nil
	}
	timestamp := op.VaaTimestamp

	if op.VaaIsDuplicated == nil {
		return nil
	}
	isDuplicated := *op.VaaIsDuplicated

	return &VaaDto{
		ID:               messageID,
		Version:          version,
		EmitterChain:     emitterChainID,
		EmitterAddr:      emitterAddr,
		Sequence:         sequence,
		GuardianSetIndex: guardianSetIndex,
		Vaa:              raw,
		Timestamp:        timestamp,
		UpdatedAt:        op.VaaUpdatedAt,
		IndexedAt:        op.VaaCreatedAt,
		IsDuplicated:     isDuplicated,
	}
}

func (r *PostgresRepository) toPayload(op *operationResult) (*map[string]any, error) {

	var payload map[string]any
	if op.PropertiesPayload == nil {
		return nil, nil
	}
	err := json.Unmarshal(*op.PropertiesPayload, &payload)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

func (r *PostgresRepository) toStandardizedProperties(op *operationResult) (*StandardizedProperties, error) {
	var properties StandardizedProperties
	if op.PropertiesRawStandardFields == nil {
		return nil, nil
	}
	err := json.Unmarshal(*op.PropertiesRawStandardFields, &properties)
	if err != nil {
		return nil, err
	}
	return &properties, nil
}

func (r *PostgresRepository) toNestedSourceTx(op *operationResult) (*AttributeDoc, error) {

	var attribute *AttributeDoc
	if op.TransactionType == "nested-source-tx" {
		var denormalizedNestedTxHash string
		chainID := sdk.ChainID(op.TransactionChainID)
		if op.TransactionTxHash != "" {
			// denormalize tx hash for compatibility reasons.
			denormalizedNestedTxHash = domain.DenormalizeTxHashByChainId(chainID, op.TransactionTxHash)
		}
		values := map[string]any{
			"originChainId": op.TransactionChainID,
			"originTxHash":  denormalizedNestedTxHash,
		}
		if op.TransactionFromAddress != nil && *op.TransactionFromAddress != "" {
			// denormalize from address for compatibility reasons.
			denormalizedOriginAddress := domain.DenormalizeAddressByChainId(
				chainID, *op.TransactionFromAddress)
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

type operationPrices struct {
	Symbol     string
	TotalUSD   string
	TotalToken string
}

func (r *PostgresRepository) toPrices(op *operationResult) *operationPrices {

	var priceSymbol, priceUsdAmount, priceTokenAmount string

	if op.PriceSymbol != nil {
		priceSymbol = *op.PriceSymbol
	}
	if op.PriceTotalUSD != nil {
		priceUsdAmount = op.PriceTotalUSD.String()
	}
	if op.PriceTotalToken != nil {
		priceTokenAmount = op.PriceTotalToken.String()
	}

	return &operationPrices{
		Symbol:     priceSymbol,
		TotalUSD:   priceUsdAmount,
		TotalToken: priceTokenAmount,
	}
}
