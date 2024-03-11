package protocols_test

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	contributorsHandlerPkg "github.com/wormhole-foundation/wormhole-explorer/api/handlers/protocols"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/protocols"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

func TestGetContributorsTotalValues(t *testing.T) {

	app := fiber.New()
	defer app.Shutdown()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})

	input := []struct {
		testName             string
		mockError            string
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			testName:             "succeed scenario",
			mockError:            "",
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: "[{\"protocol\":\"protocol1\",\"total_messages\":0}]",
		},
		{
			testName:             "fail scenario",
			mockError:            errors.New("mock_error").Error(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "[{\"protocol\":\"protocol1\",\"total_messages\":0,\"error\":\"mock_error\"}]",
		},
	}

	for _, inputArgs := range input {
		t.Run(inputArgs.testName, func(t *testing.T) {

			service := mockService(func(ctx context.Context) []contributorsHandlerPkg.ProtocolTotalValuesDTO {
				return []contributorsHandlerPkg.ProtocolTotalValuesDTO{
					{
						ProtocolStats: contributorsHandlerPkg.ProtocolStats{
							Protocol: "protocol1",
						},
						Error: inputArgs.mockError,
					},
				}
			})

			controller := protocols.NewController(zap.NewNop(), service)
			err := controller.GetProtocolsTotalValues(c)
			assert.Nil(t, err)
			assert.Equal(t, inputArgs.expectedStatusCode, c.Response().StatusCode())
			assert.Equal(t, inputArgs.expectedResponseBody, string(c.Response().Body()))
		})
	}

}

type mockService func(ctx context.Context) []contributorsHandlerPkg.ProtocolTotalValuesDTO

func (m mockService) GetProtocolsTotalValues(ctx context.Context) []contributorsHandlerPkg.ProtocolTotalValuesDTO {
	return m(ctx)
}
