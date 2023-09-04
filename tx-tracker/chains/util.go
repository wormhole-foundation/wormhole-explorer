package chains

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
)

// timestampFromHex converts a hex timestamp into a `time.Time` value.
func timestampFromHex(s string) (time.Time, error) {

	// remove the leading "0x" or "0X" from the hex string
	hexDigits := strings.Replace(s, "0x", "", 1)
	hexDigits = strings.Replace(hexDigits, "0X", "", 1)

	// parse the hex digits into an integer
	epoch, err := strconv.ParseInt(hexDigits, 16, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse hex timestamp: %w", err)
	}

	// convert the unix epoch into a `time.Time` value
	timestamp := time.Unix(epoch, 0).UTC()
	return timestamp, nil
}

// httpGet is a helper function that performs an HTTP request.
func httpGet(ctx context.Context, rateLimiter *time.Ticker, url string) ([]byte, error) {

	// Wait for the rate limiter
	if !waitForRateLimiter(ctx, rateLimiter) {
		return nil, ctx.Err()
	}

	// Build the HTTP request
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send it
	var client http.Client
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to query url: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", response.StatusCode)
	}

	// Read the response body and return
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

// httpPost is a helper function that performs an HTTP request.
func httpPost(ctx context.Context, rateLimiter *time.Ticker, url string, body any) ([]byte, error) {

	// Wait for the rate limiter
	if !waitForRateLimiter(ctx, rateLimiter) {
		return nil, ctx.Err()
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Build the HTTP request
	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	// Send it
	var client http.Client
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to query url: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", response.StatusCode)
	}

	// Read the response body and return
	result, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return result, nil
}

func waitForRateLimiter(ctx context.Context, t *time.Ticker) bool {
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// rateLimitedRpcClient is a wrapper around `rpc.Client` that adds rate limits
type rateLimitedRpcClient struct {
	client *rpc.Client
}

func rpcDialContext(ctx context.Context, url string) (*rateLimitedRpcClient, error) {

	client, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}

	tmp := rateLimitedRpcClient{
		client: client,
	}
	return &tmp, nil
}

func (c *rateLimitedRpcClient) CallContext(
	ctx context.Context,
	rateLimiter *time.Ticker,
	result interface{},
	method string,
	args ...interface{},
) error {

	if !waitForRateLimiter(ctx, rateLimiter) {
		return ctx.Err()
	}

	return c.client.CallContext(ctx, result, method, args...)
}

func (c *rateLimitedRpcClient) Close() {
	c.client.Close()
}

func txHashLowerCaseWith0x(v string) string {
	if strings.HasPrefix(v, "0x") {
		return strings.ToLower(v)
	}
	return "0x" + strings.ToLower(v)
}
