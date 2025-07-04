# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go client library for the Robinhood crypto trading API. The project has been migrated from `samjtro/go-robinhoodcrypto` to `rizome-dev/go-robinhood` and refactored to use `pkg/crypto` structure for broader API coverage.

## Development Commands

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./pkg/auth
go test ./pkg/client
```

### Building
```bash
# Build all packages
go build ./...

# Build and run examples
go run ./internal/examples/basic/main.go
go run ./internal/examples/orders/main.go
go run ./internal/examples/pagination/main.go
go run ./internal/examples/crypto_list/main.go
go run ./internal/examples/advanced/main.go
```

### Linting and Formatting
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run golangci-lint if available
golangci-lint run
```

### Module Management
```bash
# Update dependencies
go mod tidy

# Download dependencies
go mod download

# Verify dependencies
go mod verify
```

## Architecture

### Core Package Structure
- `pkg/crypto/auth/` - Ed25519 authentication for Robinhood API
- `pkg/crypto/client/` - Main client with HTTP handling, rate limiting, and service clients
- `pkg/crypto/models/` - Data models for API requests/responses
- `pkg/crypto/errors/` - Error handling and API error parsing
- `pkg/crypto/ratelimit/` - Rate limiting implementation (100 req/min, 300 burst)

### Service Architecture
The client follows a service-oriented architecture:
- **AccountService** - Account details and crypto trading account info
- **MarketDataService** - Best bid/ask prices, estimated prices
- **TradingService** - Trading pairs, holdings, orders (get/place/cancel)

### Authentication
Uses Ed25519 digital signatures with:
- API key from Robinhood (`x-api-key`)
- Timestamp (`x-timestamp`)
- Message signature (`x-signature`)

### Key Features
- Automatic rate limiting (100 requests/minute, 300 burst)
- Automatic retries for transient failures
- Pagination support for large result sets
- Comprehensive error handling with typed API errors
- Support for all major order types (market, limit, stop loss, stop limit)

## Important Implementation Details

### Rate Limiting
The client implements token bucket rate limiting that respects Robinhood's limits:
- 100 requests per minute sustained
- 300 requests per minute burst capacity
- Automatic backoff on 429 responses

### Error Handling
API errors are parsed into structured `APIError` types with:
- HTTP status code
- Error type and message
- Field-specific validation errors
- Proper error chaining

### Order Management
- Client order IDs are auto-generated UUIDs if not provided
- All order types support proper validation
- Order states are tracked through the full lifecycle

## Example Usage Patterns

### Basic Client Setup
```go
import "github.com/rizome-dev/go-robinhood/pkg/crypto/client"

client, err := client.New(apiKey, privateKey)
// With custom options
client, err := client.New(apiKey, privateKey, 
    client.WithHTTPClient(httpClient),
    client.WithRateLimiter(rateLimiter))
```

### Error Handling Pattern
```go
import "github.com/rizome-dev/go-robinhood/pkg/crypto/errors"

if err != nil {
    if apiErr, ok := err.(*errors.APIError); ok {
        // Handle structured API errors
        fmt.Printf("API Error: %s\n", apiErr.Type)
    }
    return err
}
```

### Pagination Pattern
```go
import "github.com/rizome-dev/go-robinhood/pkg/crypto/client"

paginator := client.Trading.NewOrdersPaginator(filter)
for paginator.HasNext() {
    orders, err := paginator.Next(ctx)
    // Process orders
}
```

## Migration Notes

This codebase has been migrated from `samjtro/go-robinhoodcrypto` to `rizome-dev/go-robinhood` with the following changes:
- Module name updated to `github.com/rizome-dev/go-robinhood`
- Library moved from `pkg/` to `pkg/crypto/` for broader API coverage
- All import paths updated throughout codebase
- Examples and documentation updated to reflect new structure

## Environment Setup

Set these environment variables for testing:
```bash
export ROBINHOOD_API_KEY="rh-api-your-key-here"
export ROBINHOOD_PRIVATE_KEY="your-base64-private-key"
```

## Testing Strategy

- Unit tests for all major components
- Integration tests using real API credentials (when available)
- Mock-based testing for HTTP client interactions
- Race condition testing for concurrent operations
- Coverage testing to ensure comprehensive test coverage

## Security Considerations

- Private keys are never logged or exposed
- API keys are handled securely
- All requests use HTTPS
- Rate limiting prevents API abuse
- Proper input validation on all user inputs