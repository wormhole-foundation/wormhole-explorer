package solana

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"go.uber.org/ratelimit"
)

var (
	ErrTooManyRequests      = errors.New("too many requests")
	ErrSlotSkippedOrMissing = errors.New("slot was skipped, or missing in long-term storage")
)

type SolanaSDK struct {
	rpcClient  *rpc.Client
	commitment rpc.CommitmentType
	rl         ratelimit.Limiter
}

type GetBlockResult struct {
	IsConfirmed  bool
	Transactions []rpc.TransactionWithMeta
	BlockTime    *time.Time
}

func NewSolanaSDK(url string, rl ratelimit.Limiter) *SolanaSDK {
	return &SolanaSDK{
		rpcClient:  rpc.New(url),
		commitment: rpc.CommitmentConfirmed,
		rl:         rl,
	}
}

func (s *SolanaSDK) GetLatestBlock(ctx context.Context) (uint64, error) {
	s.rl.Take()
	return s.rpcClient.GetSlot(ctx, s.commitment)
}

func (s *SolanaSDK) GetBlock(ctx context.Context, block uint64) (*GetBlockResult, error) {
	s.rl.Take()
	rewards := false
	maxSupportedTransactionVersion := uint64(0)
	out, err := s.rpcClient.GetBlockWithOpts(ctx, block, &rpc.GetBlockOpts{
		Encoding:                       solana.EncodingBase64, // solana-go doesn't support json encoding.
		TransactionDetails:             "full",
		Rewards:                        &rewards,
		Commitment:                     s.commitment,
		MaxSupportedTransactionVersion: &maxSupportedTransactionVersion,
	})
	if err != nil {
		return nil, s.convertError(err)
	}
	if out == nil {
		// Per the API, nil just means the block is not confirmed.
		return &GetBlockResult{IsConfirmed: false}, nil
	}
	var blockTime *time.Time
	if out.BlockTime != nil {
		t := out.BlockTime.Time()
		blockTime = &t
	}
	return &GetBlockResult{IsConfirmed: true, Transactions: out.Transactions, BlockTime: blockTime}, nil
}

func (s *SolanaSDK) GetSignaturesForAddress(ctx context.Context, address solana.PublicKey) ([]*rpc.TransactionSignature, error) {
	s.rl.Take()
	return s.rpcClient.GetSignaturesForAddress(ctx, address)
}

func (s *SolanaSDK) GetTransaction(ctx context.Context, txSignature solana.Signature) (*rpc.GetTransactionResult, error) {
	s.rl.Take()
	maxSupportedTransactionVersion := uint64(0)
	return s.rpcClient.GetTransaction(ctx, txSignature, &rpc.GetTransactionOpts{
		Encoding:                       solana.EncodingBase64,
		Commitment:                     s.commitment,
		MaxSupportedTransactionVersion: &maxSupportedTransactionVersion,
	})
}

func (s *SolanaSDK) convertError(er error) error {

	switch err := er.(type) {
	case *jsonrpc.RPCError:
		switch err.Code {
		case http.StatusTooManyRequests:
			return ErrTooManyRequests
		case -32009:
			return ErrSlotSkippedOrMissing
		default:
			return er
		}
	default:
		return er
	}
}
