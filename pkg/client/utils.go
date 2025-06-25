package client

import (
	"context"
	"fmt"
)

// GetAllTradeablePairs fetches all tradeable cryptocurrency pairs
func (c *Client) GetAllTradeablePairs(ctx context.Context) ([]TradingPairInfo, error) {
	// Get all trading pairs (no symbols = all pairs)
	response, err := c.Trading.GetTradingPairs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get trading pairs: %w", err)
	}

	var pairs []TradingPairInfo
	for _, pair := range response.Results {
		// Only include tradable pairs
		if pair.Status == "tradable" {
			pairs = append(pairs, TradingPairInfo{
				Symbol:         pair.Symbol,
				AssetCode:      pair.AssetCode,
				QuoteCode:      pair.QuoteCode,
				Status:         pair.Status,
				MinOrderSize:   pair.MinOrderSize,
				MaxOrderSize:   pair.MaxOrderSize,
				AssetIncrement: pair.AssetIncrement,
				QuoteIncrement: pair.QuoteIncrement,
			})
		}
	}

	// TODO: Handle pagination if response.Next is not empty
	// This would require updating GetTradingPairs to accept pagination parameters

	return pairs, nil
}

// GetAllTradeableSymbols returns just the symbols of all tradeable pairs
func (c *Client) GetAllTradeableSymbols(ctx context.Context) ([]string, error) {
	pairs, err := c.GetAllTradeablePairs(ctx)
	if err != nil {
		return nil, err
	}

	symbols := make([]string, len(pairs))
	for i, pair := range pairs {
		symbols[i] = pair.Symbol
	}

	return symbols, nil
}

// TradingPairInfo contains information about a trading pair
type TradingPairInfo struct {
	Symbol         string
	AssetCode      string
	QuoteCode      string
	Status         string
	MinOrderSize   string
	MaxOrderSize   string
	AssetIncrement string
	QuoteIncrement string
}