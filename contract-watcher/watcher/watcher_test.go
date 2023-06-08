package watcher

import (
	"context"
	"testing"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
)

// TestCheckTxShouldBeUpdated tests the checkTxShouldBeUpdated function.
func TestCheckTxShouldBeUpdated(t *testing.T) {

	testCase := []struct {
		name                              string
		inputTx                           storage.TransactionUpdate
		inputGetGlobalTransactionByIDFunc FuncGetGlobalTransactionById
		expectedUpdate                    bool
		expectedError                     error
	}{
		{
			name: "tx with status completed and does not exist transaction with the same vaa ID",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusConfirmed,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{}, storage.ErrDocNotFound
			},
			expectedUpdate: true,
			expectedError:  nil,
		},
		{
			name: "tx with status completed and already exists a transaction with the same vaa ID with status completed",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusConfirmed,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{
					Destination: storage.DestinationTx{
						Status: domain.TxStatusConfirmed,
					}}, nil
			},
			expectedUpdate: true,
			expectedError:  nil,
		},
		{
			name: "tx with status completed and already exist a transaction with the same vaa ID with status failed",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusConfirmed,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{
					Destination: storage.DestinationTx{
						Status: domain.TxStatusFailedToProcess,
					}}, nil
			},
			expectedUpdate: true,
			expectedError:  nil,
		},
		{
			name: "tx with status completed and already exist a transaction with the same vaa ID with status unknown",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusConfirmed,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{
					Destination: storage.DestinationTx{
						Status: domain.TxStatusUnkonwn,
					}}, nil
			},
			expectedUpdate: true,
			expectedError:  nil,
		},
		{
			name: "tx with status failed and does not exist transaction with the same vaa ID",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusFailedToProcess,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{}, storage.ErrDocNotFound
			},
			expectedUpdate: true,
			expectedError:  nil,
		},
		{
			name: "tx with status failed and already exists a transaction with the same vaa ID with status completed",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusFailedToProcess,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{
					Destination: storage.DestinationTx{
						Status: domain.TxStatusConfirmed,
					}}, nil
			},
			expectedUpdate: false,
			expectedError:  ErrTxfailedCannotBeUpdated,
		},
		{
			name: "tx with status failed and already exist a transaction with the same vaa ID with status failed",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusFailedToProcess,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{
					Destination: storage.DestinationTx{
						Status: domain.TxStatusFailedToProcess,
					}}, nil
			},
			expectedUpdate: true,
			expectedError:  nil,
		},
		{
			name: "tx with status failed and already exist a transaction with the same vaa ID with status unknown",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusFailedToProcess,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{
					Destination: storage.DestinationTx{
						Status: domain.TxStatusUnkonwn,
					}}, nil
			},
			expectedUpdate: true,
			expectedError:  nil,
		},
		{
			name: "tx with status unknown and does not exist transaction with the same vaa ID",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusUnkonwn,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{}, storage.ErrDocNotFound
			},
			expectedUpdate: true,
			expectedError:  nil,
		},
		{
			name: "tx with status unknown and already exists a transaction with the same vaa ID with status completed",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusUnkonwn,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{
					Destination: storage.DestinationTx{
						Status: domain.TxStatusConfirmed,
					}}, nil
			},
			expectedUpdate: false,
			expectedError:  ErrTxUnknowCannotBeUpdated,
		},
		{
			name: "tx with status unknown and already exist a transaction with the same vaa ID with status failed",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusUnkonwn,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{
					Destination: storage.DestinationTx{
						Status: domain.TxStatusFailedToProcess,
					}}, nil
			},
			expectedUpdate: false,
			expectedError:  ErrTxUnknowCannotBeUpdated,
		},
		{
			name: "tx with status unknown and already exist a transaction with the same vaa ID with status unknown",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: domain.TxStatusUnkonwn,
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{
					Destination: storage.DestinationTx{
						Status: domain.TxStatusUnkonwn,
					}}, nil
			},
			expectedUpdate: true,
			expectedError:  nil,
		},
		{
			name: "tx with invalid status",
			inputTx: storage.TransactionUpdate{
				Destination: storage.DestinationTx{
					Status: "invalid_status",
				}},
			inputGetGlobalTransactionByIDFunc: func(ctx context.Context, id string) (storage.TransactionUpdate, error) {
				return storage.TransactionUpdate{}, storage.ErrDocNotFound
			},
			expectedUpdate: true,
			expectedError:  ErrInvalidTxStatus,
		},
	}

	// iterate over the test cases
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			update, err := checkTxShouldBeUpdated(context.Background(), tc.inputTx, tc.inputGetGlobalTransactionByIDFunc)
			if update != tc.expectedUpdate {
				t.Errorf("expected update %v, got %v", tc.expectedUpdate, update)
			}
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}
