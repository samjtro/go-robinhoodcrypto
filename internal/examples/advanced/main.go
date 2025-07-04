package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rizome-dev/go-robinhood/pkg/crypto/auth"
	"github.com/rizome-dev/go-robinhood/pkg/crypto/client"
	"github.com/rizome-dev/go-robinhood/pkg/crypto/errors"
	"github.com/rizome-dev/go-robinhood/pkg/crypto/models"
	"github.com/rizome-dev/go-robinhood/pkg/crypto/ratelimit"
)


func main() {
	// Example 1: Generate a new key pair
	fmt.Println("=== Generating Key Pair ===")
	privateKey, publicKey, err := auth.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}
	fmt.Printf("Private Key: %s...\n", privateKey[:20])
	fmt.Printf("Public Key: %s...\n", publicKey[:20])
	fmt.Println("Use the public key to create API credentials on Robinhood")

	// Get actual credentials from environment
	apiKey := os.Getenv("ROBINHOOD_API_KEY")
	actualPrivateKey := os.Getenv("ROBINHOOD_PRIVATE_KEY")

	if apiKey == "" || actualPrivateKey == "" {
		log.Fatal("Please set ROBINHOOD_API_KEY and ROBINHOOD_PRIVATE_KEY environment variables")
	}

	// Example 2: Create client with custom configuration
	fmt.Println("\n=== Custom Client Configuration ===")
	
	// Custom HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	
	// Custom rate limiter with conservative settings
	rateLimiter := ratelimit.NewRateLimiter(50, 50, time.Minute)
	
	// Create client with custom options
	c, err := client.New(
		apiKey, 
		actualPrivateKey,
		client.WithHTTPClient(httpClient),
		client.WithRateLimiter(rateLimiter),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 3: Error handling with different UUID scenarios
	fmt.Println("\n=== Error Handling ===")
	
	// Scenario 1: Invalid order with manual invalid UUID
	invalidOrder := &models.PlaceOrderRequest{
		Symbol:        "INVALID-SYMBOL",
		ClientOrderID: "not-a-uuid", // Manually set invalid UUID
		Side:          "buy",
		Type:          "market",
		MarketOrderConfig: &models.MarketOrderConfig{
			AssetQuantity: -1, // Negative quantity
		},
	}

	_, err = c.Trading.PlaceOrder(ctx, invalidOrder)
	if err != nil {
		// Check if it's an API error
		if apiErr, ok := err.(*errors.APIError); ok {
			fmt.Printf("API Error Type: %s\n", apiErr.Type)
			fmt.Printf("Status Code: %d\n", apiErr.StatusCode)
			for _, e := range apiErr.Errors {
				fmt.Printf("  Field: %s, Error: %s\n", e.Attr, e.Detail)
			}
		} else {
			fmt.Printf("Other error: %v\n", err)
		}
	}

	// Example 4: Concurrent operations with proper rate limiting
	fmt.Println("\n=== Concurrent Operations ===")
	
	symbols := []string{"BTC-USD", "ETH-USD", "DOGE-USD", "SOL-USD", "MATIC-USD"}
	results := make(chan string, len(symbols))
	
	// Launch concurrent price checks
	for _, symbol := range symbols {
		go func(sym string) {
			price, err := c.MarketData.GetBestBidAsk(ctx, sym)
			if err != nil {
				results <- fmt.Sprintf("%s: Error - %v", sym, err)
				return
			}
			if len(price.Results) > 0 {
				results <- fmt.Sprintf("%s: $%.2f", sym, price.Results[0].Price)
			}
		}(symbol)
	}
	
	// Collect results
	for i := 0; i < len(symbols); i++ {
		fmt.Println(<-results)
	}

	// Example 5: Monitoring order execution with auto-generated UUID
	fmt.Println("\n=== Order Monitoring ===")
	
	// Place a limit order well above market price (unlikely to fill immediately)
	bidAsk, err := c.MarketData.GetBestBidAsk(ctx, "BTC-USD")
	if err != nil {
		log.Printf("Failed to get current price: %v", err)
		return
	}
	
	currentPrice := bidAsk.Results[0].Price
	limitPrice := currentPrice * 0.90 // 10% below market (buy order)
	
	// Demonstrating auto-generated UUID - no ClientOrderID field
	monitorOrder := &models.PlaceOrderRequest{
		Symbol: "BTC-USD",
		// ClientOrderID omitted - will be auto-generated
		Side: "buy",
		Type: "limit",
		LimitOrderConfig: &models.LimitOrderConfig{
			AssetQuantity: 0.0001,
			LimitPrice:    limitPrice,
			TimeInForce:   "gtc",
		},
	}
	
	order, err := c.Trading.PlaceOrder(ctx, monitorOrder)
	if err != nil {
		log.Printf("Failed to place monitor order: %v", err)
		return
	}
	
	fmt.Printf("Placed limit order %s at $%.2f\n", order.ID[:8], limitPrice)
	fmt.Println("Monitoring order status...")
	
	// Monitor for 30 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	timeout := time.After(30 * time.Second)
	
	for {
		select {
		case <-ticker.C:
			status, err := c.Trading.GetOrder(ctx, order.ID)
			if err != nil {
				log.Printf("Failed to check order: %v", err)
				continue
			}
			
			fmt.Printf("  Status: %s", status.State)
			if status.FilledAssetQuantity > 0 {
				fmt.Printf(" (Filled: %.8f @ $%.2f)", 
					status.FilledAssetQuantity, 
					status.AveragePrice)
			}
			fmt.Println()
			
			if status.State == "filled" || status.State == "canceled" {
				fmt.Println("Order completed!")
				goto cleanup
			}
			
		case <-timeout:
			fmt.Println("Monitoring timeout reached")
			goto cleanup
		}
	}
	
cleanup:
	// Cancel the order if it's still open
	status, err := c.Trading.GetOrder(ctx, order.ID)
	if err == nil && status.State == "open" {
		fmt.Println("Cancelling open order...")
		err = c.Trading.CancelOrder(ctx, order.ID)
		if err != nil {
			log.Printf("Failed to cancel order: %v", err)
		} else {
			fmt.Println("Order cancelled successfully")
		}
	}

	// Example 6: Demonstrating manual UUID generation for tracking
	fmt.Println("\n=== Manual UUID for Tracking ===")
	
	// Sometimes you want to generate your own UUID for tracking purposes
	trackingUUID := uuid.New().String()
	fmt.Printf("Generated tracking UUID: %s\n", trackingUUID)
	
	// Use it in an order
	trackedOrder := &models.PlaceOrderRequest{
		Symbol:        "ETH-USD",
		ClientOrderID: trackingUUID, // Manual UUID for tracking
		Side:          "buy",
		Type:          "market",
		MarketOrderConfig: &models.MarketOrderConfig{
			AssetQuantity: 0.001,
		},
	}
	
	// You could save this UUID to a database for tracking
	fmt.Printf("Order prepared with tracking UUID: %s\n", trackedOrder.ClientOrderID)

	// Example 7: Using context for cancellation
	fmt.Println("\n=== Context Cancellation ===")
	
	// Create a context with timeout
	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// This will be cancelled if it takes too long
	holdings, err := c.Trading.GetHoldings(ctx2)
	if err != nil {
		if err == context.DeadlineExceeded {
			fmt.Println("Request timed out")
		} else {
			fmt.Printf("Request failed: %v\n", err)
		}
	} else {
		fmt.Printf("Retrieved %d holdings within timeout\n", len(holdings.Results))
	}
}