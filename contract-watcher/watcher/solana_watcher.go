package watcher

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/avast/retry-go"
	solana_types "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/near/borsh-go"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/solana"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type instructionID byte

const (
	unknownInstructionID           instructionID = 0x0
	completeNativeInstructionID    instructionID = 0x2
	completeWrappedInstructionID   instructionID = 0x3
	transferInstructionNumAccounts               = 14
)

const (
	unknownInstruction         = "unknown"
	completeNativeInstruction  = "completeNative"
	completeWrappedInstruction = "completeWrapped"
)

func (i instructionID) Name() string {
	switch i {
	case unknownInstructionID:
		return unknownInstruction
	case completeNativeInstructionID:
		return completeNativeInstruction
	case completeWrappedInstructionID:
		return completeWrappedInstruction
	default:
		return unknownInstruction
	}
}

const (
	postVAAAccountIndex = 2
	postVAATypeIndex    = 0x2
)

const maxRetries = 10
const retryDelay = 5 * time.Second

type SolanaWatcher struct {
	client          *solana.SolanaSDK
	chainID         vaa.ChainID
	blockchain      string
	contractAddress solana_types.PublicKey
	sizeBlocks      uint8
	waitSeconds     uint16
	initialBlock    int64
	repository      *storage.Repository
	logger          *zap.Logger
	close           chan bool
	wg              sync.WaitGroup
}
type SolanaParams struct {
	Blockchain      string
	ContractAddress solana_types.PublicKey
	SizeBlocks      uint8
	WaitSeconds     uint16
	InitialBlock    int64
}

type postVAAData struct {
	Version          uint8
	GuardianSetIndex uint32
	Timestamp        uint32
	Nonce            uint32
	EmitterChain     uint16
	EmitterAddress   [32]uint8
	Sequence         uint64
	ConsistencyLevel uint8
	Payload          []uint8
}

func (p *postVAAData) MessageID() string {
	vaa := vaa.VAA{
		EmitterChain:   vaa.ChainID(p.EmitterChain),
		EmitterAddress: vaa.Address(p.EmitterAddress),
		Sequence:       p.Sequence,
	}
	return vaa.MessageID()
}

func NewSolanaWatcher(client *solana.SolanaSDK, repo *storage.Repository, params SolanaParams, logger *zap.Logger) *SolanaWatcher {
	return &SolanaWatcher{
		client:          client,
		chainID:         vaa.ChainIDSolana,
		blockchain:      params.Blockchain,
		contractAddress: params.ContractAddress,
		sizeBlocks:      params.SizeBlocks,
		waitSeconds:     params.WaitSeconds,
		initialBlock:    params.InitialBlock,
		repository:      repo,
		logger:          logger.With(zap.String("blockchain", params.Blockchain), zap.Uint16("chainId", uint16(vaa.ChainIDSolana))),
	}
}

func (w *SolanaWatcher) Start(ctx context.Context) error {
	// get the current block for the chain.
	cBlock, err := w.repository.GetCurrentBlock(ctx, w.blockchain, w.initialBlock)
	if err != nil {
		w.logger.Error("cannot get current block", zap.Error(err))
		return err
	}
	currentBlock := uint64(cBlock)
	w.wg.Add(1)
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("clossing watcher by context")
			w.wg.Done()
			return nil
		case <-w.close:
			w.logger.Info("clossing watcher")
			w.wg.Done()
			return nil
		default:
			// get the latest block for the chain.
			lastBlock, err := w.client.GetLatestBlock(ctx)
			if err != nil {
				w.logger.Error("cannot get blockchain stats", zap.Error(err))
			}
			maxBlocks := uint64(w.sizeBlocks)
			w.logger.Info("current block", zap.Uint64("current", currentBlock), zap.Uint64("last", lastBlock))
			if currentBlock < lastBlock {
				totalBlocks := (lastBlock-currentBlock)/maxBlocks + 1
				for i := 0; i < int(totalBlocks); i++ {
					fromBlock := currentBlock + uint64(i)*maxBlocks
					toBlock := fromBlock + maxBlocks - 1
					if toBlock > lastBlock {
						toBlock = lastBlock
					}
					w.logger.Info("processing blocks", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
					w.processBlock(ctx, fromBlock, toBlock)
					w.logger.Info("blocks processed", zap.Uint64("from", fromBlock), zap.Uint64("to", toBlock))
				}
				// process all the blocks between current and last block.
			} else {
				w.logger.Info("waiting for new blocks")
				select {
				case <-ctx.Done():
					w.wg.Done()
					return nil
				case <-time.After(time.Duration(w.waitSeconds) * time.Second):
				}
			}
			currentBlock = lastBlock
		}
	}
}

func (w *SolanaWatcher) Close() {
	close(w.close)
	w.wg.Wait()
}

func (w *SolanaWatcher) processBlock(ctx context.Context, fromBlock uint64, toBlock uint64) {

	for block := fromBlock; block <= toBlock; block++ {
		w.logger.Debug("processing block", zap.Uint64("block", block))
		retry.Do(
			func() error {
				// get the transactions for the block.
				result, err := w.client.GetBlock(ctx, block)
				if err != nil {
					w.logger.Error("cannot get block", zap.Uint64("block", block), zap.Error(err))
					return nil
				}
				// check if the block is confirmed.
				if !result.IsConfirmed {
					return errors.New("block not confirmed")
				}
				for txNum, txRpc := range result.Transactions {
					w.processTransaction(ctx, txRpc, block, txNum, result.BlockTime)
				}
				// update the last block number processed in the database.
				watcherBlock := storage.WatcherBlock{
					ID:          w.blockchain,
					BlockNumber: int64(block),
					UpdatedAt:   time.Now(),
				}
				return w.repository.UpdateWatcherBlock(ctx, watcherBlock)
			},
			retry.Attempts(maxRetries),
			retry.Delay(retryDelay),
		)
	}
}

func (w *SolanaWatcher) processTransaction(ctx context.Context, txRpc rpc.TransactionWithMeta, block uint64, txNum int, blockTime *time.Time) {
	if txRpc.Meta.Err != nil {
		w.logger.Debug("Transaction failed, skipping it",
			zap.Uint64("block", block),
			zap.Int("txNum", txNum),
			zap.String("err", fmt.Sprint(txRpc.Meta.Err)),
		)
		return
	}
	tx, err := txRpc.GetTransaction()
	if err != nil {
		w.logger.Error("failed to unmarshal transaction",
			zap.Uint64("block", block),
			zap.Int("txNum", txNum),
			zap.Int("dataLen", len(txRpc.Transaction.GetBinary())),
			zap.Error(err),
		)
		return
	}
	txSignature := tx.Signatures[0]
	programIndex := int16(-1)
	for n, key := range tx.Message.AccountKeys {
		if key.Equals(w.contractAddress) {
			programIndex = int16(n)
		}
	}
	if programIndex == -1 {
		return
	}
	if txRpc.Meta.Err != nil {
		w.logger.Debug("skipping failed Wormhole transaction",
			zap.Stringer("txSignature", txSignature),
			zap.Uint64("block", block),
			zap.Error(err))
		return
	}

	w.logger.Debug("found Wormhole transaction",
		zap.Stringer("txSignature", txSignature),
		zap.Uint64("block", block))

	// Find top-level instructions
	for _, inst := range tx.Message.Instructions {
		instruccionID := getTransferInstruction(inst, programIndex)
		switch instruccionID {
		case completeNativeInstructionID, completeWrappedInstructionID:
		default:
			continue
		}

		found, accountAddress := w.getAccountAddress(inst, programIndex, tx)
		if !found || accountAddress == nil {
			continue
		}

		signaturesByAccount, err := w.client.GetSignaturesForAddress(ctx, *accountAddress)
		if err != nil {
			w.logger.Error("getting signatures for address failed",
				zap.Stringer("txSignature", txSignature),
				zap.Uint64("block", block),
				zap.String("accountAddress", accountAddress.String()),
				zap.Error(err))
			continue
		}

		for _, signatureByAccount := range signaturesByAccount {
			txSignatureAccount := signatureByAccount.Signature
			if !txSignature.Equals(txSignatureAccount) {
				log := w.logger.With(
					zap.Stringer("txSignature", txSignature),
					zap.Stringer("txSignatureAccount", txSignatureAccount),
					zap.Uint64("block", block),
					zap.String("accountAddress", accountAddress.String()),
				)

				result, err := w.client.GetTransaction(ctx, txSignatureAccount)
				if err != nil {
					log.Error("getting transaction failed", zap.Error(err))
					continue
				}

				if result.Transaction == nil {
					log.Error("transaction not found")
					continue
				}

				t, err := result.Transaction.GetTransaction()
				if err != nil {
					log.Error("getting transaction detail failed", zap.Error(err))
					continue
				}
				if len(t.Message.Instructions) == 1 {
					instruccion := t.Message.Instructions[0]
					if len(instruccion.Data) == 0 {
						log.Error("instruction data is empty")
						continue
					}

					if instruccion.Data[0] != postVAATypeIndex {
						log.Error("invalid instruction data type", zap.Uint8("type", uint8(instruccion.Data[0])))
						continue
					}

					var data postVAAData
					if err := borsh.Deserialize(&data, instruccion.Data[1:]); err != nil {
						log.Error("failed to deserialize instruction data", zap.Error(err))
						continue
					}

					updatedAt := time.Now()
					globalTx := storage.TransactionUpdate{
						ID: data.MessageID(),
						Destination: storage.DestinationTx{
							ChainID:     w.chainID,
							Status:      TxStatusConfirmed,
							Method:      instruccionID.Name(),
							TxHash:      txSignature.String(),
							BlockNumber: strconv.FormatUint(block, 10),
							Timestamp:   blockTime,
							UpdatedAt:   &updatedAt,
						},
					}
					err = w.repository.UpsertGlobalTransaction(ctx, globalTx)
					if err != nil {
						log.Error("cannot save redeemed tx", zap.Error(err))
					} else {
						log.Info("saved redeemed tx", zap.String("vaa", data.MessageID()))
					}
				} else {
					log.Warn("transaction has more than one instruction")
				}
			}
		}
	}
}

func getTransferInstruction(inst solana_types.CompiledInstruction, programIndex int16) instructionID {
	if inst.ProgramIDIndex != uint16(programIndex) {
		return unknownInstructionID
	}

	if len(inst.Data) == 0 {
		return unknownInstructionID
	}

	switch inst.Data[0] {
	case byte(completeNativeInstructionID):
		return completeNativeInstructionID
	case byte(completeWrappedInstructionID):
		return completeWrappedInstructionID
	default:
		return unknownInstructionID
	}
}

func (w *SolanaWatcher) getAccountAddress(inst solana_types.CompiledInstruction, programIndex int16, tx *solana_types.Transaction) (bool, *solana_types.PublicKey) {
	if len(inst.Accounts) != transferInstructionNumAccounts || len(tx.Message.AccountKeys) < postVAAAccountIndex {
		w.logger.Error("invalid number of accounts",
			zap.Int("instructionAccounts", len(inst.Accounts)),
			zap.Int("messageAccounts", len(tx.Message.AccountKeys)),
			zap.Int("postVAAAccountIndex", postVAAAccountIndex),
			zap.Int("expectedAccounts", transferInstructionNumAccounts))
		return false, nil
	}
	accountAddress := tx.Message.AccountKeys[inst.Accounts[postVAAAccountIndex]]
	return true, &accountAddress
}
