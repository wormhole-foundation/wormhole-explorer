package stats

import (
	"context"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_convertToDecimal(t *testing.T) {

	url := "https://us-east-1-1.aws.cloud2.influxdata.com"
	token := "FQ14tMrjuumxGGPlCIQvWfX_JDLUPJDOaTXKH_t3pHNDIvN13rbbmlG0JuuWvqo15Gw_qEjRqaeZ-BnCf0VaXA=="
	cli := influxdb2.NewClient(url, token)
	logger := zap.NewExample()
	ctx := context.Background()
	repo := NewRepository(cli, "xlabs", "wormscan-24hours-mainnet-staging", logger)
	result, err := repo.GetSymbolWithAssets(ctx, TimeSpan30Days)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
