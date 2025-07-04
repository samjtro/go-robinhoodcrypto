package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rizome-dev/go-robinhood/pkg/crypto/client"
)

func main() {
	// Get credentials from environment variables
	apiKey := os.Getenv("ROBINHOOD_API_KEY")
	privateKey := os.Getenv("ROBINHOOD_PRIVATE_KEY")

	if apiKey == "" || privateKey == "" {
		log.Fatal("Please set ROBINHOOD_API_KEY and ROBINHOOD_PRIVATE_KEY environment variables")
	}

	// Create a new client
	c, err := client.New(apiKey, privateKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Get account details
	fmt.Println("=== Account Details ===")
	account, err := c.Account.GetAccountDetails(ctx)
	if err != nil {
		log.Printf("Failed to get account details: %v", err)
	} else {
		fmt.Printf("Account Number: %s\n", account.AccountNumber)
		fmt.Printf("Status: %s\n", account.Status)
		fmt.Printf("Buying Power: %s %s\n", account.BuyingPower, account.BuyingPowerCurrency)
	}

	// Example 2: Get best bid/ask prices
	fmt.Println("\n=== Market Data ===")
	bidAsk, err := c.MarketData.GetBestBidAsk(ctx, "BTC-USD", "ETH-USD")
	if err != nil {
		log.Printf("Failed to get bid/ask: %v", err)
	} else {
		for _, result := range bidAsk.Results {
			fmt.Printf("%s - Price: %.2f, Bid: %.2f, Ask: %.2f\n",
				result.Symbol,
				result.Price,
				result.BidInclusiveOfSellSpread,
				result.AskInclusiveOfBuySpread)
		}
	}

	// Example 3: Get estimated prices for different quantities
	fmt.Println("\n=== Estimated Prices ===")
	estimates, err := c.MarketData.GetEstimatedPrice(ctx, "BTC-USD", "ask", 0.001, 0.01, 0.1)
	if err != nil {
		log.Printf("Failed to get estimated prices: %v", err)
	} else {
		for _, est := range estimates.Results {
			fmt.Printf("Quantity %.4f %s: $%.2f\n", est.Quantity, est.Symbol, est.Price)
		}
	}

	// Example 4: Get trading pairs
	fmt.Println("\n=== Trading Pairs ===")
	pairs, err := c.Trading.GetTradingPairs(ctx, "BTC-USD", "ETH-USD")
	if err != nil {
		log.Printf("Failed to get trading pairs: %v", err)
	} else {
		for _, pair := range pairs.Results {
			fmt.Printf("%s: Min Order: %s, Max Order: %s, Status: %s\n",
				pair.Symbol,
				pair.MinOrderSize,
				pair.MaxOrderSize,
				pair.Status)
		}
	}

	// Example 5: Get holdings
	fmt.Println("\n=== Holdings ===")
	holdings, err := c.Trading.GetHoldings(ctx)
	if err != nil {
		log.Printf("Failed to get holdings: %v", err)
	} else {
		for _, holding := range holdings.Results {
			fmt.Printf("%s: Total: %.8f, Available: %.8f\n",
				holding.AssetCode,
				holding.TotalQuantity,
				holding.QuantityAvailableForTrading)
		}
	}

	// Example 6: Get recent orders
	fmt.Println("\n=== Recent Orders ===")
	orders, err := c.Trading.GetOrders(ctx, nil)
	if err != nil {
		log.Printf("Failed to get orders: %v", err)
	} else {
		for _, order := range orders.Results {
			fmt.Printf("Order %s: %s %s %s - Status: %s\n",
				order.ID[:8],
				order.Side,
				order.Type,
				order.Symbol,
				order.State)
		}
	}
}