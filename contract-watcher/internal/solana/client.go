package solana

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolanaSDK struct {
	rpcClient  *rpc.Client
	commitment rpc.CommitmentType
}

type GetBlockResult struct {
	IsConfirmed  bool
	Transactions []rpc.TransactionWithMeta
}

func NewSolanaSDK(url string) *SolanaSDK {
	return &SolanaSDK{
		rpcClient:  rpc.New(url),
		commitment: rpc.CommitmentConfirmed,
	}
}

func (s *SolanaSDK) GetLatestBlock(ctx context.Context) (uint64, error) {
	return s.rpcClient.GetSlot(ctx, s.commitment)
}

func (s *SolanaSDK) GetBlock(ctx context.Context, block uint64) (*GetBlockResult, error) {
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
		return nil, err
	}
	if out == nil {
		// Per the API, nil just means the block is not confirmed.
		return &GetBlockResult{IsConfirmed: false}, nil
	}
	return &GetBlockResult{IsConfirmed: true, Transactions: out.Transactions}, nil
}

func (s *SolanaSDK) GetSignaturesForAddress(ctx context.Context, address solana.PublicKey) ([]*rpc.TransactionSignature, error) {
	return s.rpcClient.GetSignaturesForAddress(ctx, address)
}

func (s *SolanaSDK) GetTransaction(ctx context.Context, txSignature solana.Signature) (*rpc.GetTransactionResult, error) {
	maxSupportedTransactionVersion := uint64(0)
	return s.rpcClient.GetTransaction(ctx, txSignature, &rpc.GetTransactionOpts{
		Encoding:                       solana.EncodingBase64,
		Commitment:                     s.commitment,
		MaxSupportedTransactionVersion: &maxSupportedTransactionVersion,
	})
}
