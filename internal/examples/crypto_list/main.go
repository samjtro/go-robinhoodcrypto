package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/samjtro/go-robinhoodcrypto/pkg/client"
	"github.com/samjtro/go-robinhoodcrypto/pkg/models"
)

func main() {
	// Get API credentials from environment
	apiKey := os.Getenv("ROBINHOOD_API_KEY")
	privateKey := os.Getenv("ROBINHOOD_PRIVATE_KEY")

	if apiKey == "" || privateKey == "" {
		log.Fatal("Please set ROBINHOOD_API_KEY and ROBINHOOD_PRIVATE_KEY environment variables")
	}

	// Create client
	c, err := client.New(apiKey, privateKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Get all tradeable cryptocurrencies
	fmt.Println("Fetching all tradeable cryptocurrencies from Robinhood...")
	cryptos, err := c.GetAllTradeableCryptos(ctx)
	if err != nil {
		log.Printf("Failed to fetch crypto list: %v", err)
	} else {
		fmt.Printf("Found %d tradeable cryptocurrencies:\n", len(cryptos))
		for i, crypto := range cryptos {
			if i < 10 { // Show first 10
				fmt.Printf("  %s (%s)\n", crypto.Symbol, crypto.AssetCode)
			}
		}
		if len(cryptos) > 10 {
			fmt.Printf("  ... and %d more\n", len(cryptos)-10)
		}
	}

	// Example 2: Get just the symbols
	fmt.Println("\nFetching crypto symbols...")
	symbols, err := c.GetAllTradeableCryptoSymbols(ctx)
	if err != nil {
		log.Printf("Failed to fetch crypto symbols: %v", err)
	} else {
		fmt.Printf("Found %d symbols: %v...\n", len(symbols), symbols[:min(5, len(symbols))])
	}

	// Example 3: Place an order with auto-generated UUID
	fmt.Println("\nExample: Creating order request with auto-generated UUID...")
	
	// Create order request without ClientOrderID
	orderReq := &models.PlaceOrderRequest{
		Symbol: "BTC-USD",
		Side:   "buy",
		Type:   "market",
		MarketOrderConfig: &models.MarketOrderConfig{
			AssetQuantity: 0.0001,
		},
	}
	
	fmt.Println("Order request created without ClientOrderID - it will be auto-generated")
	fmt.Printf("Symbol: %s, Side: %s, Type: %s\n", orderReq.Symbol, orderReq.Side, orderReq.Type)
	
	// Note: We're not actually placing the order in this example
	// To place the order, you would use:
	// order, err := c.Trading.PlaceOrder(ctx, orderReq)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}