package contributors_test

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	contributorsHandlerPkg "github.com/wormhole-foundation/wormhole-explorer/api/handlers/contributors"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/contributors"
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
			expectedResponseBody: "[{\"contributor\":\"contributor1\",\"total_messages\":\"\",\"last_day_messages\":\"\",\"last_day_diff_percentage\":\"\"}]",
		},
		{
			testName:             "fail scenario",
			mockError:            errors.New("mock_error").Error(),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "[{\"contributor\":\"contributor1\",\"total_messages\":\"\",\"last_day_messages\":\"\",\"last_day_diff_percentage\":\"\",\"error\":\"mock_error\"}]",
		},
	}

	for _, inputArgs := range input {
		t.Run(inputArgs.testName, func(t *testing.T) {

			service := mockService(func(ctx context.Context) []contributorsHandlerPkg.ContributorTotalValuesDTO {
				return []contributorsHandlerPkg.ContributorTotalValuesDTO{
					{
						Contributor: "contributor1",
						Error:       inputArgs.mockError,
					},
				}
			})

			controller := contributors.NewController(zap.NewNop(), service)
			err := controller.GetContributorsTotalValues(c)
			assert.Nil(t, err)
			assert.Equal(t, inputArgs.expectedStatusCode, c.Response().StatusCode())
			assert.Equal(t, inputArgs.expectedResponseBody, string(c.Response().Body()))
		})
	}

}

type mockService func(ctx context.Context) []contributorsHandlerPkg.ContributorTotalValuesDTO

func (m mockService) GetContributorsTotalValues(ctx context.Context) []contributorsHandlerPkg.ContributorTotalValuesDTO {
	return m(ctx)
}
