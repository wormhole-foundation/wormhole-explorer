package consumer_test

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/consumer"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/queue"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"sync"
	"testing"
	"time"
)

type testCase struct {
	name           string
	mockRepository func(*mockSQLRepository)
	snsPublish     func(ctx context.Context, message topic.SnsMessage) error
	eventData      *queue.Event
	expectedFails  int
	expectedDones  int
}

func TestConsumer_Start(t *testing.T) {
	testCases := []testCase{
		{
			name: "GetTxHash failed",
			mockRepository: func(mockSQL *mockSQLRepository) {
				mockSQL.On("GetTxHash", mock.Anything, mock.Anything).Return("", errors.New("mocked_error"))
			},
			expectedFails: 1,
			expectedDones: 0,
			eventData: &queue.Event{
				ID: "vaa_digest_test",
			},
		},
		{
			name: "CreateOperationTransaction failed",
			mockRepository: func(mockSQL *mockSQLRepository) {
				mockSQL.On("GetTxHash", mock.Anything, mock.Anything).Return("tx_hash_test", nil)
				mockSQL.On("CreateOperationTransaction", mock.Anything, mock.Anything).Return(errors.New("mocked_error"))
			},
			eventData: &queue.Event{
				ChainID:        vaa.ChainIDEthereum,
				Type:           "source-chain-event",
				ID:             "vaa_digest_test",
				VaaId:          "vaa_id_test",
				EmitterAddress: "emitter_address_test",
				Timestamp:      &time.Time{},
			},
			expectedFails: 1,
			expectedDones: 0,
		},
		{
			name: "PublishVaa failed",
			mockRepository: func(mockSQL *mockSQLRepository) {
				mockSQL.On("GetTxHash", mock.Anything, mock.Anything).Return("tx_hash_test", nil)
				mockSQL.On("CreateOperationTransaction", mock.Anything, mock.Anything).Return(nil)
			},
			eventData: &queue.Event{
				ChainID:        vaa.ChainIDEthereum,
				Type:           "source-chain-event",
				ID:             "vaa_digest_test",
				VaaId:          "vaa_id_test",
				EmitterAddress: "emitter_address_test",
				Timestamp:      &time.Time{},
			},
			snsPublish: func(ctx context.Context, message topic.SnsMessage) error {
				return errors.New("mocked_error")
			},
			expectedFails: 1,
			expectedDones: 0,
		},
		{
			name: "Process Success",
			mockRepository: func(mockSQL *mockSQLRepository) {
				mockSQL.On("GetTxHash", mock.Anything, mock.Anything).Return("tx_hash_test", nil)
				mockSQL.On("CreateOperationTransaction", mock.Anything, mock.Anything).Return(nil)
			},
			eventData: &queue.Event{
				ChainID:        vaa.ChainIDEthereum,
				Type:           "source-chain-event",
				ID:             "vaa_digest_test",
				VaaId:          "vaa_id_test",
				EmitterAddress: "emitter_address_test",
				Timestamp:      &time.Time{},
			},
			snsPublish: func(ctx context.Context, message topic.SnsMessage) error {
				return nil
			},
			expectedFails: 0,
			expectedDones: 1,
		},
		{
			name: "Process Wormchain",
			mockRepository: func(mockSQL *mockSQLRepository) {
				mockSQL.On("GetTxHash", mock.Anything, mock.Anything).Return("tx_hash_test", nil)
			},
			eventData: &queue.Event{
				ChainID:        vaa.ChainIDWormchain,
				Type:           "source-chain-event",
				ID:             "vaa_digest_test",
				VaaId:          "vaa_id_test",
				EmitterAddress: "emitter_address_test",
				Timestamp:      &time.Time{},
			},
			snsPublish: func(ctx context.Context, message topic.SnsMessage) error {
				return nil
			},
			expectedFails: 0,
			expectedDones: 1,
		},
		{
			name: "Process Solana",
			mockRepository: func(mockSQL *mockSQLRepository) {
				mockSQL.On("GetTxHash", mock.Anything, mock.Anything).Return("tx_hash_test", nil)
			},
			eventData: &queue.Event{
				ChainID:        vaa.ChainIDSolana,
				Type:           "source-chain-event",
				ID:             "vaa_digest_test",
				VaaId:          "vaa_id_test",
				EmitterAddress: "emitter_address_test",
				Timestamp:      &time.Time{},
			},
			snsPublish: func(ctx context.Context, message topic.SnsMessage) error {
				return nil
			},
			expectedFails: 0,
			expectedDones: 1,
		},
		{
			name: "Process Aptos",
			mockRepository: func(mockSQL *mockSQLRepository) {
				mockSQL.On("GetTxHash", mock.Anything, mock.Anything).Return("tx_hash_test", nil)
			},
			eventData: &queue.Event{
				ChainID:        vaa.ChainIDAptos,
				Type:           "source-chain-event",
				ID:             "vaa_digest_test",
				VaaId:          "vaa_id_test",
				EmitterAddress: "emitter_address_test",
				Timestamp:      &time.Time{},
			},
			snsPublish: func(ctx context.Context, message topic.SnsMessage) error {
				return nil
			},
			expectedFails: 0,
			expectedDones: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel() // close ctx in order to release goroutines

			mockSQL := &mockSQLRepository{}
			tc.mockRepository(mockSQL)

			wg := sync.WaitGroup{}
			wg.Add(1)
			gotFailed := 0
			gotDone := 0
			msg := &mockMsg{
				data: tc.eventData,
				failed: func() {
					gotFailed++
					wg.Done()
				},
				done: func() {
					gotDone++
					wg.Done()
				},
			}

			instance := consumer.New(mockSQL,
				logger.New("pipeline-test"),
				tc.snsPublish,
				metrics.NewDummyMetrics(),
				1)
			instance.Start(ctx, mockConsumeFunc(msg))
			wg.Wait()

			assert.Equal(t, tc.expectedFails, gotFailed, "Expected failed messages did not match")
			assert.Equal(t, tc.expectedDones, gotDone, "Expected done messages did not match")
		})
	}
}

func mockConsumeFunc(msg *mockMsg) func(context.Context) <-chan queue.ConsumerMessage {
	return func(ctx context.Context) <-chan queue.ConsumerMessage {
		ch := make(chan queue.ConsumerMessage)
		go func() {
			ch <- msg
		}()
		return ch
	}
}

type mockSQLRepository struct {
	mock.Mock
}

func (m *mockSQLRepository) GetTxHash(ctx context.Context, vaaDigest string) (string, error) {
	args := m.Called(ctx, vaaDigest)
	return args.String(0), args.Error(1)
}

func (m *mockSQLRepository) CreateOperationTransaction(ctx context.Context, opTx consumer.OperationTransaction) error {
	args := m.Called(ctx, opTx)
	return args.Error(0)
}

type mockMsg struct {
	data         *queue.Event
	done, failed func()
}

func (m *mockMsg) Retry() uint8 {
	return 0
}

func (m *mockMsg) Data() *queue.Event {
	return m.data
}

func (m *mockMsg) Done() {
	m.done()
}

func (m *mockMsg) Failed() {
	m.failed()
}

func (m *mockMsg) IsExpired() bool {
	return false
}

func (m *mockMsg) SentTimestamp() *time.Time {
	now := time.Now()
	return &now
}
