package consumer_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/consumer"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"testing"
	"time"
)

func Test_ProcessSourceTx_AlreadyProcessed(t *testing.T) {

	testCases := []struct {
		name        string
		params      *consumer.ProcessSourceTxParams
		mockMongoDB func() *mockMongoDBRepository
		mockSQL     func() *mockSqlRepository
		expectedErr error
	}{
		{
			name: "Test_ProcessSourceTx_RunModeMongodb_AlreadyProcessed_Error",
			params: &consumer.ProcessSourceTxParams{
				RunMode:   config.RunModeMongo,
				VaaId:     "vaa_id_test",
				Overwrite: false,
			},
			mockSQL: func() *mockSqlRepository {
				return nil
			},
			mockMongoDB: func() *mockMongoDBRepository {
				m := &mockMongoDBRepository{}
				m.On("AlreadyProcessed", mock.Anything, mock.Anything).Return(false, errors.New("mocked_error"))
				return m
			},
			expectedErr: errors.New("mocked_error"),
		},
		{
			name: "Test_ProcessSourceTx_RunModeMongodb_AlreadyProcessed",
			params: &consumer.ProcessSourceTxParams{
				RunMode:   config.RunModeMongo,
				VaaId:     "vaa_id_test",
				Overwrite: false,
			},
			mockSQL: func() *mockSqlRepository {
				return nil
			},
			mockMongoDB: func() *mockMongoDBRepository {
				m := &mockMongoDBRepository{}
				m.On("AlreadyProcessed", mock.Anything, mock.Anything).Return(true, nil)
				return m
			},
			expectedErr: consumer.ErrAlreadyProcessed,
		},
		{
			name: "Test_ProcessSourceTx_RunModePostgresql_AlreadyProcessed",
			params: &consumer.ProcessSourceTxParams{
				RunMode:   config.RunModePostgresql,
				VaaId:     "vaa_id_test",
				ID:        "vaa_digest",
				Overwrite: false,
			},
			mockMongoDB: func() *mockMongoDBRepository {
				return nil
			},
			mockSQL: func() *mockSqlRepository {
				m := &mockSqlRepository{}
				m.On("AlreadyProcessed", mock.Anything, mock.Anything).Return(true, nil)
				return m
			},
			expectedErr: consumer.ErrAlreadyProcessed,
		},
		{
			name: "Test_ProcessSourceTx_RunModePostgresql_AlreadyProcessed_Error",
			params: &consumer.ProcessSourceTxParams{
				RunMode:   config.RunModePostgresql,
				VaaId:     "vaa_id_test",
				ID:        "vaa_digest",
				Overwrite: false,
			},
			mockMongoDB: func() *mockMongoDBRepository {
				return nil
			},
			mockSQL: func() *mockSqlRepository {
				m := &mockSqlRepository{}
				m.On("AlreadyProcessed", mock.Anything, mock.Anything).Return(false, errors.New("mocked_error"))
				return m
			},
			expectedErr: errors.New("mocked_error"),
		},
		{
			name: "Test_ProcessSourceTx_RunModeBoth_AlreadyProcessed",
			params: &consumer.ProcessSourceTxParams{
				RunMode:   config.RunModeBoth,
				VaaId:     "vaa_id_test",
				ID:        "vaa_digest",
				Overwrite: false,
			},
			mockMongoDB: func() *mockMongoDBRepository {
				m := &mockMongoDBRepository{}
				m.On("AlreadyProcessed", mock.Anything, mock.Anything).Return(true, nil)
				return m
			},
			mockSQL: func() *mockSqlRepository {
				m := &mockSqlRepository{}
				m.On("AlreadyProcessed", mock.Anything, mock.Anything).Return(true, nil)
				return m
			},
			expectedErr: consumer.ErrAlreadyProcessed,
		},
		{
			name: "Test_ProcessSourceTx_RunModeBoth_MongoFails",
			params: &consumer.ProcessSourceTxParams{
				RunMode:   config.RunModeBoth,
				VaaId:     "vaa_id_test",
				ID:        "vaa_digest",
				Overwrite: false,
			},
			mockMongoDB: func() *mockMongoDBRepository {
				m := &mockMongoDBRepository{}
				m.On("AlreadyProcessed", mock.Anything, mock.Anything).Return(true, errors.New("mongodb_error"))
				return m
			},
			mockSQL: func() *mockSqlRepository {
				m := &mockSqlRepository{}
				m.On("AlreadyProcessed", mock.Anything, mock.Anything).Return(true, nil)
				return m
			},
			expectedErr: errors.New("mongodb_error"),
		},
		{
			name: "Test_ProcessSourceTx_RunModeBoth_PostgresqlFails",
			params: &consumer.ProcessSourceTxParams{
				RunMode:   config.RunModeBoth,
				VaaId:     "vaa_id_test",
				ID:        "vaa_digest",
				Overwrite: false,
			},
			mockMongoDB: func() *mockMongoDBRepository {
				m := &mockMongoDBRepository{}
				m.On("AlreadyProcessed", mock.Anything, mock.Anything).Return(true, nil)
				return m
			},
			mockSQL: func() *mockSqlRepository {
				m := &mockSqlRepository{}
				m.On("AlreadyProcessed", mock.Anything, mock.Anything).Return(false, errors.New("postresql_error"))
				return m
			},
			expectedErr: errors.New("postresql_error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, gotErr := consumer.ProcessSourceTx(context.TODO(),
				nil,
				nil,
				nil,
				tc.mockMongoDB(),
				tc.params,
				"testnet",
				nil,
				tc.mockSQL())
			if gotErr.Error() != tc.expectedErr.Error() {
				t.Errorf("expected error %v, got %v", tc.expectedErr, gotErr)
			}
		})
	}
}

type mockSqlRepository struct {
	mock.Mock
}

func (m *mockSqlRepository) UpsertOriginTx(ctx context.Context, params *consumer.UpsertOriginTxParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *mockSqlRepository) UpsertTargetTx(ctx context.Context, globalTx *consumer.TargetTxUpdate) error {
	args := m.Called(ctx, globalTx)
	return args.Error(0)
}

func (m *mockSqlRepository) GetTxStatus(ctx context.Context, targetTxUpdate *consumer.TargetTxUpdate) (string, error) {
	args := m.Called(ctx, targetTxUpdate)
	return args.String(0), args.Error(1)
}

func (m *mockSqlRepository) AlreadyProcessed(ctx context.Context, vaDigest string) (bool, error) {
	args := m.Called(ctx, vaDigest)
	return args.Bool(0), args.Error(1)
}

func (m *mockSqlRepository) RegisterProcessedVaa(ctx context.Context, vaaDigest, vaaId string) error {
	args := m.Called(ctx, vaaDigest, vaaId)
	return args.Error(0)
}

type mockMongoDBRepository struct {
	mock.Mock
}

func (m *mockMongoDBRepository) UpsertOriginTx(ctx context.Context, params *consumer.UpsertOriginTxParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *mockMongoDBRepository) AlreadyProcessed(ctx context.Context, vaaId string) (bool, error) {
	args := m.Called(ctx, vaaId)
	return args.Bool(0), args.Error(1)
}

func (m *mockMongoDBRepository) GetVaaIdTxHash(ctx context.Context, id string) (*consumer.VaaIdTxHash, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*consumer.VaaIdTxHash), args.Error(1)
}

func (m *mockMongoDBRepository) UpsertTargetTx(ctx context.Context, globalTx *consumer.TargetTxUpdate) error {
	args := m.Called(ctx, globalTx)
	return args.Error(0)
}

func (m *mockMongoDBRepository) GetTxStatus(ctx context.Context, targetTxUpdate *consumer.TargetTxUpdate) (string, error) {
	args := m.Called(ctx, targetTxUpdate)
	return args.String(0), args.Error(1)
}

func (m *mockMongoDBRepository) CountDocumentsByVaas(ctx context.Context, emitterChainID sdk.ChainID, emitterAddress string, sequence string) (uint64, error) {
	args := m.Called(ctx, emitterChainID, emitterAddress, sequence)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockMongoDBRepository) GetDocumentsByVaas(ctx context.Context, lastId string, lastTimestamp *time.Time, limit uint, emitterChainID sdk.ChainID, emitterAddress string, sequence string) ([]consumer.GlobalTransaction, error) {
	args := m.Called(ctx, lastId, lastTimestamp, limit, emitterChainID, emitterAddress, sequence)
	return args.Get(0).([]consumer.GlobalTransaction), args.Error(1)
}

func (m *mockMongoDBRepository) FindSourceTxById(ctx context.Context, id string) (*consumer.SourceTxDoc, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*consumer.SourceTxDoc), args.Error(1)
}
