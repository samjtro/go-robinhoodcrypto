package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// CryptoListItem represents a crypto item from the Robinhood list
type CryptoListItem struct {
	Symbol      string
	Name        string
	AssetCode   string
}

// GetAllTradeableCryptos fetches all tradeable cryptocurrencies from Robinhood's public list
func (c *Client) GetAllTradeableCryptos(ctx context.Context) ([]CryptoListItem, error) {
	// Create a new HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Fetch the webpage
	req, err := http.NewRequestWithContext(ctx, "GET", "https://robinhood.com/lists/robinhood/97b746a5-bc2f-4c64-a828-1af0fc399bf9", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crypto list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the HTML to extract crypto data
	cryptos, err := parseRobinhoodCryptoList(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse crypto list: %w", err)
	}

	return cryptos, nil
}

// parseRobinhoodCryptoList parses the HTML content to extract crypto information
func parseRobinhoodCryptoList(html string) ([]CryptoListItem, error) {
	var cryptos []CryptoListItem

	// Look for JSON data embedded in the page
	// Robinhood typically embeds data in script tags
	jsonPattern := regexp.MustCompile(`window\.__PRELOADED_STATE__\s*=\s*({.*?});`)
	matches := jsonPattern.FindStringSubmatch(html)
	
	if len(matches) > 1 {
		// Parse the JSON data
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(matches[1]), &data); err == nil {
			// Navigate through the JSON structure to find crypto list
			if cryptos := extractCryptosFromJSON(data); len(cryptos) > 0 {
				return cryptos, nil
			}
		}
	}

	// Fallback: Look for crypto symbols in the HTML
	// This regex looks for patterns like "BTC-USD", "ETH-USD", etc.
	symbolPattern := regexp.MustCompile(`([A-Z]{2,10})-USD`)
	symbolMatches := symbolPattern.FindAllStringSubmatch(html, -1)
	
	seen := make(map[string]bool)
	for _, match := range symbolMatches {
		if len(match) > 1 {
			assetCode := match[1]
			symbol := match[0]
			
			// Skip duplicates
			if seen[symbol] {
				continue
			}
			seen[symbol] = true
			
			// Skip common false positives
			if assetCode == "USD" || assetCode == "US" {
				continue
			}
			
			cryptos = append(cryptos, CryptoListItem{
				Symbol:    symbol,
				AssetCode: assetCode,
				Name:      assetCode, // Name would need to be fetched separately
			})
		}
	}

	if len(cryptos) == 0 {
		return nil, fmt.Errorf("no cryptocurrencies found in the list")
	}

	return cryptos, nil
}

// extractCryptosFromJSON attempts to extract crypto data from parsed JSON
func extractCryptosFromJSON(data map[string]interface{}) []CryptoListItem {
	var cryptos []CryptoListItem
	
	// This is a simplified extraction - the actual structure may vary
	// You might need to adjust this based on the actual JSON structure
	var traverse func(v interface{})
	traverse = func(v interface{}) {
		switch val := v.(type) {
		case map[string]interface{}:
			// Look for crypto-related fields
			if symbol, ok := val["symbol"].(string); ok && strings.HasSuffix(symbol, "-USD") {
				crypto := CryptoListItem{
					Symbol:    symbol,
					AssetCode: strings.TrimSuffix(symbol, "-USD"),
				}
				if name, ok := val["name"].(string); ok {
					crypto.Name = name
				}
				cryptos = append(cryptos, crypto)
			}
			// Recursively traverse
			for _, v := range val {
				traverse(v)
			}
		case []interface{}:
			for _, item := range val {
				traverse(item)
			}
		}
	}
	
	traverse(data)
	return cryptos
}

// GetAllTradeableCryptoSymbols is a convenience function that returns just the symbols
func (c *Client) GetAllTradeableCryptoSymbols(ctx context.Context) ([]string, error) {
	cryptos, err := c.GetAllTradeableCryptos(ctx)
	if err != nil {
		return nil, err
	}

	symbols := make([]string, len(cryptos))
	for i, crypto := range cryptos {
		symbols[i] = crypto.Symbol
	}
	
	return symbols, nil
}