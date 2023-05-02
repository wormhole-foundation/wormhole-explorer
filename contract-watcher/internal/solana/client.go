package solana

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"go.uber.org/ratelimit"
)

// https://github.com/solana-labs/solana/blob/master/rpc-client-api/src/custom_error.rs
const (
	//Slot was skipped, or missing due to ledger jump to recent snapshot
	ErrSlotSkippedCode = -32007

	//Slot was skipped, or missing in long-term storage
	ErrLongTermStorageSlotSkippedCode = -32009
)

var (
	ErrTooManyRequests = errors.New("too many requests")
	ErrSlotSkipped     = errors.New("slot was skipped")
)

type SolanaSDK struct {
	rpcClient  *rpc.Client
	commitment rpc.CommitmentType
	rl         ratelimit.Limiter
	retries    uint
	delay      time.Duration
}

type GetBlockResult struct {
	IsConfirmed  bool
	Transactions []rpc.TransactionWithMeta
	BlockTime    *time.Time
}

type Options func(*SolanaSDK)

func NewSolanaSDK(url string, rl ratelimit.Limiter, opts ...Options) *SolanaSDK {
	r := &SolanaSDK{
		rpcClient:  rpc.New(url),
		commitment: rpc.CommitmentConfirmed,
		rl:         rl,
		retries:    0,
		delay:      0,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func WithRetries(retries uint, delay time.Duration) Options {
	return func(s *SolanaSDK) {
		s.retries = retries
		s.delay = delay
	}
}

func (s *SolanaSDK) GetLatestBlock(ctx context.Context) (uint64, error) {
	s.rl.Take()
	var slot uint64
	err := s.withRetry(func() error {
		var er error
		slot, er = s.rpcClient.GetSlot(ctx, s.commitment)
		return s.convertError(er)
	})
	return slot, err
}

func (s *SolanaSDK) GetBlock(ctx context.Context, block uint64) (*GetBlockResult, error) {
	s.rl.Take()
	rewards := false
	maxSupportedTransactionVersion := uint64(0)
	var result *GetBlockResult
	err := s.withRetry(func() error {
		out, er := s.rpcClient.GetBlockWithOpts(ctx, block, &rpc.GetBlockOpts{
			Encoding:                       solana.EncodingBase64, // solana-go doesn't support json encoding.
			TransactionDetails:             "full",
			Rewards:                        &rewards,
			Commitment:                     s.commitment,
			MaxSupportedTransactionVersion: &maxSupportedTransactionVersion,
		})
		if er != nil {
			return s.convertError(er)
		}
		if out == nil {
			// Per the API, nil just means the block is not confirmed.
			result = &GetBlockResult{IsConfirmed: false}
			return nil
		}
		var blockTime *time.Time
		if out.BlockTime != nil {
			t := out.BlockTime.Time()
			blockTime = &t
		}
		result = &GetBlockResult{IsConfirmed: true, Transactions: out.Transactions, BlockTime: blockTime}
		return nil
	})
	return result, err
}

func (s *SolanaSDK) GetSignaturesForAddress(ctx context.Context, address solana.PublicKey) ([]*rpc.TransactionSignature, error) {
	s.rl.Take()
	var result []*rpc.TransactionSignature
	err := s.withRetry(func() error {
		var er error
		result, er = s.rpcClient.GetSignaturesForAddress(ctx, address)
		return er
	})
	return result, err
}

func (s *SolanaSDK) GetTransaction(ctx context.Context, txSignature solana.Signature) (*rpc.GetTransactionResult, error) {
	s.rl.Take()
	maxSupportedTransactionVersion := uint64(0)
	var result *rpc.GetTransactionResult
	err := s.withRetry(func() error {
		var er error
		result, er = s.rpcClient.GetTransaction(ctx, txSignature, &rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase64,
			Commitment:                     s.commitment,
			MaxSupportedTransactionVersion: &maxSupportedTransactionVersion,
		})
		return s.convertError(er)
	})
	return result, err
}

func (s *SolanaSDK) convertError(er error) error {

	switch err := er.(type) {
	case *jsonrpc.RPCError:
		switch err.Code {
		case http.StatusTooManyRequests:
			return ErrTooManyRequests
		case ErrSlotSkippedCode, ErrLongTermStorageSlotSkippedCode:
			return ErrSlotSkipped
		default:
			return er
		}
	default:
		return er
	}
}

func (s *SolanaSDK) withRetry(fn func() error) error {
	return retry.Do(
		func() error {
			return fn()
		},
		retry.Attempts(s.retries),
		retry.Delay(s.delay),
		retry.RetryIf(
			func(err error) bool {
				return err == ErrTooManyRequests
			},
		),
		retry.LastErrorOnly(true),
	)
}
