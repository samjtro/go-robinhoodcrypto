package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rizome-dev/go-robinhood/pkg/crypto/auth"
	"github.com/rizome-dev/go-robinhood/pkg/crypto/errors"
	"github.com/rizome-dev/go-robinhood/pkg/crypto/models"
	"github.com/rizome-dev/go-robinhood/pkg/crypto/ratelimit"
)

const (
	defaultBaseURL = "https://trading.robinhood.com"
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
	retryDelay     = time.Second
)

// Client is the main client for interacting with the Robinhood Crypto API
type Client struct {
	httpClient    *http.Client
	baseURL       string
	auth          *auth.Authenticator
	rateLimiter   *ratelimit.RateLimiter
	
	// Service clients
	Account    *AccountService
	MarketData *MarketDataService
	Trading    *TradingService
}

// Option is a functional option for configuring the client
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL sets a custom base URL
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithRateLimiter sets a custom rate limiter
func WithRateLimiter(rl *ratelimit.RateLimiter) Option {
	return func(c *Client) {
		c.rateLimiter = rl
	}
}

// New creates a new Robinhood Crypto API client
func New(apiKey, privateKey string, opts ...Option) (*Client, error) {
	authenticator, err := auth.NewAuthenticator(apiKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	c := &Client{
		httpClient:  &http.Client{Timeout: defaultTimeout},
		baseURL:     defaultBaseURL,
		auth:        authenticator,
		rateLimiter: ratelimit.DefaultRateLimiter(),
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Initialize service clients
	c.Account = &AccountService{client: c}
	c.MarketData = &MarketDataService{client: c}
	c.Trading = &TradingService{client: c}

	return c, nil
}

// request performs an HTTP request with authentication and rate limiting
func (c *Client) request(ctx context.Context, method, path string, query url.Values, body interface{}) (*http.Response, error) {
	// Build URL
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	// Prepare body
	var bodyReader io.Reader
	var bodyStr string
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
		bodyStr = string(bodyBytes)
	}

	// Retry loop
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt)):
			}
		}

		// Rate limiting
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		
		// Add authentication headers
		// Include query parameters in the path for signature generation
		pathWithQuery := u.Path
		if u.RawQuery != "" {
			pathWithQuery = u.Path + "?" + u.RawQuery
		}
		authHeaders, err := c.auth.GetAuthHeaders(method, pathWithQuery, bodyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to get auth headers: %w", err)
		}
		for k, v := range authHeaders {
			req.Header.Set(k, v)
		}

		// Perform request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Check for rate limit errors
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			lastErr = fmt.Errorf("rate limited (429)")
			continue
		}

		// Check for server errors that should be retried
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error (%d)", resp.StatusCode)
			continue
		}

		// Success or client error (don't retry)
		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// do performs a request and handles the response
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, result interface{}) error {
	resp, err := c.request(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.ParseAPIError(respBody, resp.StatusCode)
	}

	// Parse successful response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// AccountService handles account-related endpoints
type AccountService struct {
	client *Client
}

// GetAccountDetails fetches the crypto trading account details
func (s *AccountService) GetAccountDetails(ctx context.Context) (*models.AccountDetails, error) {
	var result models.AccountDetails
	err := s.client.do(ctx, "GET", "/api/v1/crypto/trading/accounts/", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// MarketDataService handles market data endpoints
type MarketDataService struct {
	client *Client
}

// GetBestBidAsk fetches the best bid and ask prices for the given symbols
func (s *MarketDataService) GetBestBidAsk(ctx context.Context, symbols ...string) (*models.BestBidAskResponse, error) {
	query := url.Values{}
	for _, symbol := range symbols {
		query.Add("symbol", strings.ToUpper(symbol))
	}

	var result models.BestBidAskResponse
	err := s.client.do(ctx, "GET", "/api/v1/crypto/marketdata/best_bid_ask/", query, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetEstimatedPrice fetches estimated prices for different quantities
func (s *MarketDataService) GetEstimatedPrice(ctx context.Context, symbol, side string, quantities ...float64) (*models.EstimatedPriceResponse, error) {
	if side != "bid" && side != "ask" && side != "both" {
		return nil, fmt.Errorf("invalid side: must be 'bid', 'ask', or 'both'")
	}

	quantityStrs := make([]string, len(quantities))
	for i, q := range quantities {
		quantityStrs[i] = fmt.Sprintf("%g", q)
	}

	query := url.Values{
		"symbol":   []string{strings.ToUpper(symbol)},
		"side":     []string{side},
		"quantity": []string{strings.Join(quantityStrs, ",")},
	}

	var result models.EstimatedPriceResponse
	err := s.client.do(ctx, "GET", "/api/v1/crypto/marketdata/estimated_price/", query, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// TradingService handles trading-related endpoints
type TradingService struct {
	client *Client
}