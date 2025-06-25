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

	// Example 1: Get all tradeable cryptocurrency pairs with full details
	fmt.Println("=== Fetching All Tradeable Cryptocurrency Pairs ===")
	pairs, err := c.GetAllTradeablePairs(ctx)
	if err != nil {
		log.Printf("Failed to fetch tradeable pairs: %v", err)
	} else {
		fmt.Printf("Found %d tradeable cryptocurrency pairs:\n\n", len(pairs))
		
		// Show first 10 with details
		for i, pair := range pairs {
			if i < 10 {
				fmt.Printf("%2d. %s\n", i+1, pair.Symbol)
				fmt.Printf("    Asset: %s, Quote: %s\n", pair.AssetCode, pair.QuoteCode)
				fmt.Printf("    Order Size: Min %s, Max %s\n", pair.MinOrderSize, pair.MaxOrderSize)
				fmt.Printf("    Increments: Asset %s, Quote %s\n", pair.AssetIncrement, pair.QuoteIncrement)
				fmt.Println()
			}
		}
		
		if len(pairs) > 10 {
			fmt.Printf("... and %d more pairs\n\n", len(pairs)-10)
		}
	}

	// Example 2: Get just the symbols
	fmt.Println("=== Getting Symbol List ===")
	symbols, err := c.GetAllTradeableSymbols(ctx)
	if err != nil {
		log.Printf("Failed to fetch symbols: %v", err)
	} else {
		fmt.Printf("Tradeable symbols (%d total):\n", len(symbols))
		// Print symbols in columns
		for i := 0; i < len(symbols); i += 5 {
			for j := 0; j < 5 && i+j < len(symbols); j++ {
				fmt.Printf("%-12s", symbols[i+j])
			}
			fmt.Println()
			if i >= 15 { // Show first 20 symbols
				fmt.Printf("... and %d more\n", len(symbols)-20)
				break
			}
		}
	}

	// Example 3: Get specific trading pairs details
	fmt.Println("\n=== Getting Specific Trading Pairs ===")
	specificPairs, err := c.Trading.GetTradingPairs(ctx, "BTC-USD", "ETH-USD", "DOGE-USD")
	if err != nil {
		log.Printf("Failed to fetch specific pairs: %v", err)
	} else {
		fmt.Printf("Details for specific pairs:\n")
		for _, pair := range specificPairs.Results {
			fmt.Printf("\n%s (Status: %s)\n", pair.Symbol, pair.Status)
			fmt.Printf("  Min Order: %s %s\n", pair.MinOrderSize, pair.AssetCode)
			fmt.Printf("  Max Order: %s %s\n", pair.MaxOrderSize, pair.AssetCode)
			fmt.Printf("  Price Increment: %s\n", pair.QuoteIncrement)
			fmt.Printf("  Quantity Increment: %s\n", pair.AssetIncrement)
		}
	}

	// Example 4: Demonstrate order with auto-generated UUID
	fmt.Println("\n=== Order Example with Auto-Generated UUID ===")
	
	// Create order request without ClientOrderID
	orderReq := &models.PlaceOrderRequest{
		Symbol: "BTC-USD",
		// ClientOrderID omitted - will be auto-generated
		Side: "buy",
		Type: "market",
		MarketOrderConfig: &models.MarketOrderConfig{
			AssetQuantity: 0.0001,
		},
	}
	
	fmt.Println("Order request created without ClientOrderID - UUID will be auto-generated when order is placed")
	fmt.Printf("Symbol: %s, Side: %s, Type: %s, Quantity: %g BTC\n", 
		orderReq.Symbol, orderReq.Side, orderReq.Type, orderReq.MarketOrderConfig.AssetQuantity)
	
	// Note: We're not actually placing the order in this example
	// To place the order, you would use:
	// order, err := c.Trading.PlaceOrder(ctx, orderReq)
}