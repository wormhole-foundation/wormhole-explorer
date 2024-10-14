package transactions_test

import (
	"context"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_GetTokenSymbolActivity(t *testing.T) {

	mockRepo := new(mockRepository)
	svc := transactions.NewService(mockRepo, cache.NewDummyCacheClient(), 0, nil, metrics.NewNoOpMetrics(), zap.NewNop())

	from := time.Now().Truncate(2 * time.Hour)
	to := time.Now().Truncate(1 * time.Hour)
	tests := []struct {
		name           string
		query          transactions.TokenSymbolActivityQuery
		mockResponse   []transactions.TokenSymbolActivityResult
		mockError      error
		expectedResult transactions.TokenSymbolActivityResponse
		expectedError  error
	}{
		{
			name: "Valid Data",
			query: transactions.TokenSymbolActivityQuery{
				From:         from.Add(-time.Hour * 24),
				To:           to,
				TokenSymbols: []string{"USDC", "USDT"},
				Timespan:     transactions.Hour,
			},
			mockResponse: []transactions.TokenSymbolActivityResult{
				{
					Symbol:           "USDC",
					EmitterChain:     1,
					DestinationChain: 2,
					Txs:              10,
					Volume:           1000.0,
					From:             from,
					To:               to,
				},
				{
					Symbol:           "USDC",
					EmitterChain:     5,
					DestinationChain: 3,
					Txs:              10,
					Volume:           1000.0,
					From:             from,
					To:               to,
				},
				{
					Symbol:           "USDT",
					EmitterChain:     1,
					DestinationChain: 3,
					Txs:              5,
					Volume:           500.0,
					From:             from,
					To:               to,
				},
			},
			mockError: nil,
			expectedResult: transactions.TokenSymbolActivityResponse{
				Tokens: []transactions.TokenSymbolActivity{
					{
						TokenSymbol:           "USDC",
						TotalMessages:         20,
						TotalValueTransferred: 2000.0,
						TimeRangeData: []*transactions.TimeRangeData[*transactions.TokenSymbolPerChainPairData]{
							{
								From:                  from,
								To:                    to,
								TotalMessages:         20,
								TotalValueTransferred: 2000.0,
								Aggregations: []*transactions.TokenSymbolPerChainPairData{
									{
										SourceChain:           1,
										TargetChain:           2,
										TotalMessages:         10,
										TotalValueTransferred: 1000.0,
									},
									{
										SourceChain:           5,
										TargetChain:           3,
										TotalMessages:         10,
										TotalValueTransferred: 1000.0,
									},
								},
							},
						},
					},
					{
						TokenSymbol:           "USDT",
						TotalMessages:         5,
						TotalValueTransferred: 500.0,
						TimeRangeData: []*transactions.TimeRangeData[*transactions.TokenSymbolPerChainPairData]{
							{
								From:                  from,
								To:                    to,
								TotalMessages:         5,
								TotalValueTransferred: 500.0,
								Aggregations: []*transactions.TokenSymbolPerChainPairData{
									{
										SourceChain:           1,
										TargetChain:           3,
										TotalMessages:         5,
										TotalValueTransferred: 500.0,
									},
								},
							},
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "No Data",
			query: transactions.TokenSymbolActivityQuery{
				From:         from.Add(-time.Hour * 24),
				To:           from,
				TokenSymbols: []string{"TOKEN3"},
			},
			mockResponse:   []transactions.TokenSymbolActivityResult{},
			mockError:      nil,
			expectedResult: transactions.TokenSymbolActivityResponse{Tokens: []transactions.TokenSymbolActivity{}},
			expectedError:  nil,
		},
		{
			name: "Error from Repository",
			query: transactions.TokenSymbolActivityQuery{
				From:         from.Add(-time.Hour * 24),
				To:           from,
				TokenSymbols: []string{"TOKEN1"},
			},
			mockResponse:   nil,
			mockError:      assert.AnError,
			expectedResult: transactions.TokenSymbolActivityResponse{},
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.On("FindTokenSymbolActivity", mock.Anything, tt.query).Return(tt.mockResponse, tt.mockError)
			result, err := svc.GetTokenSymbolActivity(context.Background(), tt.query)
			assert.Equal(t, tt.expectedError, err)
			if len(tt.expectedResult.Tokens) != len(result.Tokens) {
				t.Errorf("Expected %d tokens, got %d", len(tt.expectedResult.Tokens), len(result.Tokens))
			}

			for _, expectedToken := range tt.expectedResult.Tokens {
				found := false
				for _, resToken := range result.Tokens {
					if expectedToken.TokenSymbol == resToken.TokenSymbol {
						found = true
						assert.Equal(t, expectedToken.TotalMessages, resToken.TotalMessages)
						assert.Equal(t, expectedToken.TotalValueTransferred, resToken.TotalValueTransferred)
						if len(expectedToken.TimeRangeData) != len(resToken.TimeRangeData) {
							t.Errorf("Expected %d TimeRangeData, got %d", len(expectedToken.TimeRangeData), len(resToken.TimeRangeData))
						}
						for _, expectedTimeRangeData := range expectedToken.TimeRangeData {
							foundTimeRangeData := false
							for _, resTimeRangeData := range resToken.TimeRangeData {
								if expectedTimeRangeData.From == resTimeRangeData.From && expectedTimeRangeData.To == resTimeRangeData.To {
									foundTimeRangeData = true
									assert.Equal(t, expectedTimeRangeData.TotalMessages, resTimeRangeData.TotalMessages)
									assert.Equal(t, expectedTimeRangeData.TotalValueTransferred, resTimeRangeData.TotalValueTransferred)
									if len(expectedTimeRangeData.Aggregations) != len(resTimeRangeData.Aggregations) {
										t.Errorf("Expected %d Aggregations, got %d",
											len(expectedTimeRangeData.Aggregations),
											len(resTimeRangeData.Aggregations))
									}
									for _, expectedAggregation := range expectedTimeRangeData.Aggregations {
										foundAggregation := false
										for _, resAggregation := range resTimeRangeData.Aggregations {
											if expectedAggregation.SourceChain == resAggregation.SourceChain && expectedAggregation.TargetChain == resAggregation.TargetChain {
												foundAggregation = true
												assert.Equal(t, expectedAggregation.TotalMessages, resAggregation.TotalMessages)
												assert.Equal(t, expectedAggregation.TotalValueTransferred, resAggregation.TotalValueTransferred)
											}
										}
										if !foundAggregation {
											t.Errorf("Aggregation source_chain %d ; targe_chain %d not found in time range:[%s-%s] for token %s",
												expectedAggregation.SourceChain,
												expectedAggregation.TargetChain,
												expectedTimeRangeData.From.String(),
												expectedTimeRangeData.To.String(),
												expectedToken.TokenSymbol)
										}
									}
								}
							}
							if !foundTimeRangeData {
								t.Errorf("TimeRangeData from %s to %s not found for token:%s", expectedTimeRangeData.From.String(), expectedTimeRangeData.To.String(), resToken.TokenSymbol)
							}
						}
					}
				}
				if !found {
					t.Errorf("Token %s not found in result", expectedToken.TokenSymbol)
				}
			}

		})
	}
}

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetTopAssets(ctx context.Context, timeSpan *transactions.TopStatisticsTimeSpan) ([]transactions.AssetDTO, error) {
	called := m.Called(ctx, timeSpan)
	return called.Get(0).([]transactions.AssetDTO), called.Error(1)
}

func (m *mockRepository) GetTopChainPairs(ctx context.Context, timeSpan *transactions.TopStatisticsTimeSpan) ([]transactions.ChainPairDTO, error) {
	called := m.Called(ctx, timeSpan)
	return called.Get(0).([]transactions.ChainPairDTO), called.Error(1)
}

func (m *mockRepository) FindChainActivity(ctx context.Context, q *transactions.ChainActivityQuery) ([]transactions.ChainActivityResult, error) {
	called := m.Called(ctx, q)
	return called.Get(0).([]transactions.ChainActivityResult), called.Error(1)
}

func (m *mockRepository) GetScorecards(ctx context.Context, usePostgres bool) (*transactions.Scorecards, error) {
	args := m.Called(ctx, usePostgres)
	return args.Get(0).(*transactions.Scorecards), args.Error(1)
}

func (m *mockRepository) FindGlobalTransactionByID(ctx context.Context, usePostgres bool, q *transactions.GlobalTransactionQuery) (*transactions.GlobalTransactionDoc, error) {
	args := m.Called(ctx, q)
	return args.Get(0).(*transactions.GlobalTransactionDoc), args.Error(1)
}

func (m *mockRepository) FindTransactions(ctx context.Context, usePostgres bool, input *transactions.FindTransactionsInput) ([]transactions.TransactionDto, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]transactions.TransactionDto), args.Error(1)
}

func (m *mockRepository) ListTransactionsByAddress(ctx context.Context, usePostgres bool, address string, pagination *pagination.Pagination) ([]transactions.TransactionDto, error) {
	args := m.Called(ctx, usePostgres, address, pagination)
	return args.Get(0).([]transactions.TransactionDto), args.Error(1)
}

func (m *mockRepository) FindChainActivityTops(ctx *fasthttp.RequestCtx, q transactions.ChainActivityTopsQuery) ([]transactions.ChainActivityTopResult, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]transactions.ChainActivityTopResult), args.Error(1)
}

func (m *mockRepository) FindApplicationActivity(ctx *fasthttp.RequestCtx, q transactions.ApplicationActivityQuery) ([]transactions.ApplicationActivityTotalsResult, []transactions.ApplicationActivityResult, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]transactions.ApplicationActivityTotalsResult), args.Get(1).([]transactions.ApplicationActivityResult), args.Error(2)
}

func (m *mockRepository) FindTokensVolume(ctx context.Context) ([]transactions.TokenVolume, error) {
	args := m.Called(ctx)
	return args.Get(0).([]transactions.TokenVolume), args.Error(1)
}

func (m *mockRepository) GetTransactionCount(ctx context.Context, q *transactions.TransactionCountQuery) ([]transactions.TransactionCountResult, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]transactions.TransactionCountResult), args.Error(1)
}

func (m *mockRepository) FindTokenSymbolActivity(ctx context.Context, payload transactions.TokenSymbolActivityQuery) ([]transactions.TokenSymbolActivityResult, error) {
	args := m.Called(ctx, payload)
	return args.Get(0).([]transactions.TokenSymbolActivityResult), args.Error(1)
}
