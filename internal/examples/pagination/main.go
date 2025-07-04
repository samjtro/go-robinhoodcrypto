package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

	// Example 1: Paginate through all trading pairs
	fmt.Println("=== Paginating Trading Pairs ===")
	pairsPaginator := c.Trading.NewTradingPairsPaginator()
	
	pageNum := 1
	for {
		pairs, err := pairsPaginator.Next(ctx)
		if err != nil {
			log.Printf("Failed to get page %d: %v", pageNum, err)
			break
		}
		
		if len(pairs) == 0 {
			break
		}
		
		fmt.Printf("Page %d - Found %d trading pairs\n", pageNum, len(pairs))
		for _, pair := range pairs[:min(3, len(pairs))] { // Show first 3
			fmt.Printf("  %s (Status: %s)\n", pair.Symbol, pair.Status)
		}
		
		if !pairsPaginator.HasNext() {
			break
		}
		pageNum++
	}

	// Example 2: Get all holdings using pagination
	fmt.Println("\n=== Getting All Holdings ===")
	holdingsPaginator := c.Trading.NewHoldingsPaginator()
	
	allHoldings, err := holdingsPaginator.GetAllPages(ctx)
	if err != nil {
		log.Printf("Failed to get all holdings: %v", err)
	} else {
		fmt.Printf("Total holdings: %d\n", len(allHoldings))
		for _, holding := range allHoldings {
			if holding.TotalQuantity > 0 {
				fmt.Printf("  %s: %.8f\n", holding.AssetCode, holding.TotalQuantity)
			}
		}
	}

	// Example 3: Paginate through filtered orders
	fmt.Println("\n=== Paginating Recent Orders ===")
	
	// Filter for orders in the last 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	filter := &models.OrdersFilter{
		CreatedAtStart: &sevenDaysAgo,
		Limit:          10, // 10 per page
	}
	
	ordersPaginator := c.Trading.NewOrdersPaginator(filter)
	
	totalOrders := 0
	pageNum = 1
	
	for {
		orders, err := ordersPaginator.Next(ctx)
		if err != nil {
			log.Printf("Failed to get orders page %d: %v", pageNum, err)
			break
		}
		
		if len(orders) == 0 {
			break
		}
		
		fmt.Printf("Page %d - Found %d orders\n", pageNum, len(orders))
		for _, order := range orders {
			fmt.Printf("  %s: %s %s %s (State: %s)\n",
				order.ID[:8],
				order.Side,
				order.Type,
				order.Symbol,
				order.State)
		}
		
		totalOrders += len(orders)
		
		if !ordersPaginator.HasNext() {
			break
		}
		pageNum++
		
		// Limit to first 3 pages for demo
		if pageNum > 3 {
			fmt.Println("  ... (limiting to 3 pages for demo)")
			break
		}
	}
	
	fmt.Printf("\nTotal orders found: %d\n", totalOrders)

	// Example 4: Manual pagination control
	fmt.Println("\n=== Manual Pagination Control ===")
	
	// Get first page of BTC orders
	btcFilter := &models.OrdersFilter{
		Symbol: "BTC-USD",
		Limit:  5,
	}
	
	firstPage, err := c.Trading.GetOrders(ctx, btcFilter)
	if err != nil {
		log.Printf("Failed to get first page: %v", err)
	} else {
		fmt.Printf("First page has %d orders\n", len(firstPage.Results))
		
		// Check if there's a next page
		if firstPage.Next != "" {
			// Extract cursor from the next URL and fetch next page
			// This is handled automatically by the paginator, but shown here for clarity
			fmt.Println("Next page available")
		}
		
		// Check if there's a previous page
		if firstPage.Previous != "" {
			fmt.Println("Previous page available")
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}