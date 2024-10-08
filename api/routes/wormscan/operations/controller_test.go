package operations_test

import (
	"context"
	"io"
	"net/http"
	"slices"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/test-go/testify/mock"
	ops "github.com/wormhole-foundation/wormhole-explorer/api/handlers/operations"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/operations"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

func Test_FindAll(t *testing.T) {

	testCases := []struct {
		name               string
		requestURL         string
		expectedStatusCode int
		expectedResponse   string
		setupServiceMock   func(*mockOpsService)
	}{
		{
			name:               "Test_FindAll_SinglePayloadType",
			requestURL:         "/api/v1/operations?payloadType=1",
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"operations":[]}`,
			setupServiceMock: func(mockService *mockOpsService) {
				payloadTypeMatcher := mock.MatchedBy(func(filter ops.OperationFilter) bool {
					return slices.Equal(filter.PayloadType, []int{1})
				})
				mockService.On("FindAll", mock.Anything, payloadTypeMatcher).Return([]*ops.OperationDto{}, nil)
			},
		},
		{
			name:               "Test_FindAll_MultiplePayloadType",
			requestURL:         "/api/v1/operations?payloadType=1,2,3",
			expectedStatusCode: http.StatusOK,
			expectedResponse:   `{"operations":[]}`,
			setupServiceMock: func(mockService *mockOpsService) {
				payloadTypeMatcher := mock.MatchedBy(func(filter ops.OperationFilter) bool {
					return slices.Equal(filter.PayloadType, []int{1, 2, 3})
				})
				mockService.On("FindAll", mock.Anything, payloadTypeMatcher).Return([]*ops.OperationDto{}, nil)
			},
		},
		{
			name:               "Test_FindAll_InvalidPayloadType",
			requestURL:         "/api/v1/operations?payloadType=1,thisShouldBeANumber,3",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   `{"code":3,"message":"invalid payloadType","details":[{"request_id":"\u003cnil\u003e"}]}`,
			setupServiceMock:   func(mockService *mockOpsService) {},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			req, err := http.NewRequest(http.MethodGet, testCase.requestURL, nil)
			if err != nil {
				t.Fatal(err)
			}

			mockService := &mockOpsService{}
			testCase.setupServiceMock(mockService)

			app := fiber.New(fiber.Config{
				ErrorHandler:          middleware.ErrorHandler,
				DisableStartupMessage: true,
				Immutable:             true,
			})
			app.Get("/api/v1/operations", operations.NewController(mockService, zap.NewNop()).FindAll)

			resp, _ := app.Test(req, 1000)
			defer resp.Body.Close()

			if resp.StatusCode != testCase.expectedStatusCode {
				t.Fatalf("expected status code %d, got %d", testCase.expectedStatusCode, resp.StatusCode)
			}

			respBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if string(respBytes) != testCase.expectedResponse {
				t.Fatalf("expected response %s, got %s", testCase.expectedResponse, string(respBytes))
			}

		})
	}

}

type mockOpsService struct {
	mock.Mock
}

func (m *mockOpsService) FindById(ctx context.Context, usePostgres bool, chainID vaa.ChainID, emitter *types.Address, seq string) (*ops.OperationDto, error) {
	args := m.Called(ctx, usePostgres, chainID, emitter, seq)
	return args.Get(0).(*ops.OperationDto), args.Error(1)
}
func (m *mockOpsService) FindAll(ctx context.Context, filter ops.OperationFilter) ([]*ops.OperationDto, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*ops.OperationDto), args.Error(1)
}
