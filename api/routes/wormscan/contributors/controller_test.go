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
		mockError          error
		expectedStatusCode int
	}{
		{
			mockError:          nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			mockError:          errors.New("mock_error"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, inputArgs := range input {
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
	}

}

type mockService func(ctx context.Context) []contributorsHandlerPkg.ContributorTotalValuesDTO

func (m mockService) GetContributorsTotalValues(ctx context.Context) []contributorsHandlerPkg.ContributorTotalValuesDTO {
	return m(ctx)
}
