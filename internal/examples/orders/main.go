package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rizome-dev/go-robinhood/pkg/crypto/client"
	"github.com/rizome-dev/go-robinhood/pkg/crypto/models"
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

	// Example 1: Place a market buy order with auto-generated UUID
	fmt.Println("=== Placing Market Buy Order (Auto-Generated UUID) ===")
	marketOrder := &models.PlaceOrderRequest{
		Symbol: "BTC-USD",
		// ClientOrderID is not set - will be auto-generated
		Side: "buy",
		Type: "market",
		MarketOrderConfig: &models.MarketOrderConfig{
			AssetQuantity: 0.0001, // Buy 0.0001 BTC
		},
	}

	order, err := c.Trading.PlaceOrder(ctx, marketOrder)
	if err != nil {
		log.Printf("Failed to place market order: %v", err)
	} else {
		fmt.Printf("Market order placed successfully!\n")
		fmt.Printf("Order ID: %s\n", order.ID)
		fmt.Printf("Status: %s\n", order.State)
		fmt.Printf("Symbol: %s\n", order.Symbol)
	}

	// Example 2: Place a limit sell order with manual UUID
	fmt.Println("\n=== Placing Limit Sell Order (Manual UUID) ===")
	
	// First, get current price to set a reasonable limit
	bidAsk, err := c.MarketData.GetBestBidAsk(ctx, "BTC-USD")
	if err != nil {
		log.Printf("Failed to get current price: %v", err)
		return
	}
	
	currentPrice := bidAsk.Results[0].Price
	limitPrice := currentPrice * 1.01 // Set limit 1% above current price

	// You can still manually set ClientOrderID if you want to track it
	manualUUID := "custom-" + fmt.Sprintf("%.2f", currentPrice) // Example of custom ID
	limitOrder := &models.PlaceOrderRequest{
		Symbol:        "BTC-USD",
		ClientOrderID: manualUUID, // Manually set for tracking
		Side:          "sell",
		Type:          "limit",
		LimitOrderConfig: &models.LimitOrderConfig{
			AssetQuantity: 0.0001,
			LimitPrice:    limitPrice,
			TimeInForce:   "gtc", // Good Till Cancelled
		},
	}

	order2, err := c.Trading.PlaceOrder(ctx, limitOrder)
	if err != nil {
		log.Printf("Failed to place limit order: %v", err)
	} else {
		fmt.Printf("Limit order placed successfully!\n")
		fmt.Printf("Order ID: %s\n", order2.ID)
		fmt.Printf("Status: %s\n", order2.State)
		fmt.Printf("Limit Price: %.2f\n", limitPrice)
	}

	// Example 3: Place a stop loss order with auto-generated UUID
	fmt.Println("\n=== Placing Stop Loss Order (Auto-Generated UUID) ===")
	
	stopPrice := currentPrice * 0.95 // Stop loss at 5% below current price

	stopLossOrder := &models.PlaceOrderRequest{
		Symbol: "BTC-USD",
		// ClientOrderID omitted - will be auto-generated
		Side: "sell",
		Type: "stop_loss",
		StopLossOrderConfig: &models.StopLossOrderConfig{
			AssetQuantity: 0.0001,
			StopPrice:     stopPrice,
			TimeInForce:   "gtc",
		},
	}

	order3, err := c.Trading.PlaceOrder(ctx, stopLossOrder)
	if err != nil {
		log.Printf("Failed to place stop loss order: %v", err)
	} else {
		fmt.Printf("Stop loss order placed successfully!\n")
		fmt.Printf("Order ID: %s\n", order3.ID)
		fmt.Printf("Status: %s\n", order3.State)
		fmt.Printf("Stop Price: %.2f\n", stopPrice)
	}

	// Example 4: Cancel an order
	if order2 != nil && order2.State == "open" {
		fmt.Println("\n=== Cancelling Limit Order ===")
		err = c.Trading.CancelOrder(ctx, order2.ID)
		if err != nil {
			log.Printf("Failed to cancel order: %v", err)
		} else {
			fmt.Printf("Order %s cancelled successfully!\n", order2.ID)
		}
	}

	// Example 5: Check order status
	if order != nil {
		fmt.Println("\n=== Checking Order Status ===")
		updatedOrder, err := c.Trading.GetOrder(ctx, order.ID)
		if err != nil {
			log.Printf("Failed to get order status: %v", err)
		} else {
			fmt.Printf("Order %s status: %s\n", updatedOrder.ID[:8], updatedOrder.State)
			if updatedOrder.State == "filled" {
				fmt.Printf("Filled quantity: %.8f\n", updatedOrder.FilledAssetQuantity)
				fmt.Printf("Average price: %.2f\n", updatedOrder.AveragePrice)
			}
		}
	}
}

// Note: This example includes order placement which will execute real trades
// Make sure to use small quantities and understand the risks involved