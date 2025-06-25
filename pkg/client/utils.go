package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	return GetAllTradeableCryptosWithDebug(ctx, false)
}

// GetAllTradeableCryptosWithDebug fetches all tradeable cryptocurrencies with optional debugging
func GetAllTradeableCryptosWithDebug(ctx context.Context, debug bool) ([]CryptoListItem, error) {
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
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	if debug {
		fmt.Printf("Debug: Making request to %s\n", req.URL)
		fmt.Printf("Debug: Headers: %v\n", req.Header)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch crypto list: %w", err)
	}
	defer resp.Body.Close()

	if debug {
		fmt.Printf("Debug: Response status: %d\n", resp.StatusCode)
		fmt.Printf("Debug: Response headers: %v\n", resp.Header)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if debug && len(body) > 0 {
			fmt.Printf("Debug: Response body (first 500 chars): %s\n", string(body[:min(500, len(body))]))
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if debug {
		fmt.Printf("Debug: Response body size: %d bytes\n", len(body))
		// Save response to file for inspection
		debugFile := "robinhood_crypto_list_debug.html"
		if err := os.WriteFile(debugFile, body, 0644); err == nil {
			fmt.Printf("Debug: Response saved to %s\n", debugFile)
		}
	}

	// Parse the HTML to extract crypto data
	cryptos, err := parseRobinhoodCryptoListWithDebug(string(body), debug)
	if err != nil {
		return nil, fmt.Errorf("failed to parse crypto list: %w", err)
	}

	return cryptos, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseRobinhoodCryptoList parses the HTML content to extract crypto information
func parseRobinhoodCryptoList(html string) ([]CryptoListItem, error) {
	return parseRobinhoodCryptoListWithDebug(html, false)
}

// parseRobinhoodCryptoListWithDebug parses the HTML content with optional debugging
func parseRobinhoodCryptoListWithDebug(html string, debug bool) ([]CryptoListItem, error) {
	var cryptos []CryptoListItem

	if debug {
		fmt.Printf("Debug: HTML length: %d characters\n", len(html))
		fmt.Printf("Debug: First 200 chars of HTML: %s\n", html[:min(200, len(html))])
	}

	// Look for JSON data embedded in the page
	// Try multiple patterns as Robinhood might use different ones
	patterns := []string{
		`window\.__PRELOADED_STATE__\s*=\s*({.*?});`,
		`window\.__APOLLO_STATE__\s*=\s*({.*?});`,
		`<script[^>]*>window\.__NEXT_DATA__\s*=\s*({.*?})</script>`,
		`data-react-props="([^"]+)"`,
		`<script[^>]*type="application/json"[^>]*>({.*?})</script>`,
	}

	for _, pattern := range patterns {
		if debug {
			fmt.Printf("Debug: Trying pattern: %s\n", pattern)
		}
		
		jsonPattern := regexp.MustCompile(pattern)
		matches := jsonPattern.FindStringSubmatch(html)
		
		if len(matches) > 1 {
			jsonStr := matches[1]
			// Unescape HTML entities if needed
			jsonStr = strings.ReplaceAll(jsonStr, "&quot;", `"`)
			jsonStr = strings.ReplaceAll(jsonStr, "&amp;", "&")
			
			if debug {
				fmt.Printf("Debug: Found JSON data with pattern %s, length: %d\n", pattern, len(jsonStr))
				fmt.Printf("Debug: First 200 chars of JSON: %s\n", jsonStr[:min(200, len(jsonStr))])
			}
			
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &data); err == nil {
				// Navigate through the JSON structure to find crypto list
				if cryptos := extractCryptosFromJSONWithDebug(data, debug); len(cryptos) > 0 {
					return cryptos, nil
				}
			} else if debug {
				fmt.Printf("Debug: Failed to parse JSON: %v\n", err)
			}
		}
	}

	if debug {
		fmt.Println("Debug: No JSON data found, trying HTML parsing")
	}

	// Fallback: Look for crypto symbols in the HTML
	// Try multiple patterns
	symbolPatterns := []string{
		`([A-Z]{2,10})-USD`,
		`"symbol":"([A-Z]{2,10}-USD)"`,
		`data-symbol="([A-Z]{2,10}-USD)"`,
		`href="/crypto/([A-Z]{2,10})"`,
		`>([A-Z]{2,10})<.*?USD`,
	}

	seen := make(map[string]bool)
	
	for _, pattern := range symbolPatterns {
		if debug {
			fmt.Printf("Debug: Trying symbol pattern: %s\n", pattern)
		}
		
		symbolPattern := regexp.MustCompile(pattern)
		symbolMatches := symbolPattern.FindAllStringSubmatch(html, -1)
		
		if debug && len(symbolMatches) > 0 {
			fmt.Printf("Debug: Found %d matches with pattern %s\n", len(symbolMatches), pattern)
		}
		
		for _, match := range symbolMatches {
			if len(match) > 1 {
				symbol := match[1]
				
				// Handle different match formats
				if !strings.Contains(symbol, "-USD") && !strings.Contains(symbol, "USD") {
					symbol = symbol + "-USD"
				}
				
				assetCode := strings.TrimSuffix(symbol, "-USD")
				
				// Skip duplicates
				if seen[symbol] {
					continue
				}
				seen[symbol] = true
				
				// Skip common false positives
				if assetCode == "USD" || assetCode == "US" || len(assetCode) < 2 {
					continue
				}
				
				if debug {
					fmt.Printf("Debug: Found crypto: %s\n", symbol)
				}
				
				cryptos = append(cryptos, CryptoListItem{
					Symbol:    symbol,
					AssetCode: assetCode,
					Name:      assetCode, // Name would need to be fetched separately
				})
			}
		}
	}

	if len(cryptos) == 0 {
		// Last resort: look for any mention of popular cryptos
		knownCryptos := []string{"BTC", "ETH", "DOGE", "SOL", "MATIC", "AVAX", "LINK", "UNI", "AAVE", "LTC"}
		for _, crypto := range knownCryptos {
			if strings.Contains(html, crypto+"-USD") || strings.Contains(html, crypto+" ") {
				cryptos = append(cryptos, CryptoListItem{
					Symbol:    crypto + "-USD",
					AssetCode: crypto,
					Name:      crypto,
				})
			}
		}
	}

	if debug {
		fmt.Printf("Debug: Total cryptos found: %d\n", len(cryptos))
	}

	if len(cryptos) == 0 {
		return nil, fmt.Errorf("no cryptocurrencies found in the list")
	}

	return cryptos, nil
}

// extractCryptosFromJSON attempts to extract crypto data from parsed JSON
func extractCryptosFromJSON(data map[string]interface{}) []CryptoListItem {
	return extractCryptosFromJSONWithDebug(data, false)
}

// extractCryptosFromJSONWithDebug attempts to extract crypto data with optional debugging
func extractCryptosFromJSONWithDebug(data map[string]interface{}, debug bool) []CryptoListItem {
	var cryptos []CryptoListItem
	seen := make(map[string]bool)
	
	if debug {
		fmt.Println("Debug: Starting JSON traversal")
	}
	
	// This is a simplified extraction - the actual structure may vary
	var traverse func(v interface{}, path string)
	traverse = func(v interface{}, path string) {
		switch val := v.(type) {
		case map[string]interface{}:
			// Look for crypto-related fields
			symbol := ""
			name := ""
			ticker := ""
			
			// Try different field names
			if s, ok := val["symbol"].(string); ok {
				symbol = s
			} else if s, ok := val["ticker"].(string); ok {
				ticker = s
			} else if s, ok := val["code"].(string); ok {
				ticker = s
			}
			
			if n, ok := val["name"].(string); ok {
				name = n
			} else if n, ok := val["display_name"].(string); ok {
				name = n
			} else if n, ok := val["full_name"].(string); ok {
				name = n
			}
			
			// Check if this looks like crypto data
			if symbol != "" && (strings.HasSuffix(symbol, "-USD") || strings.HasSuffix(symbol, "USD")) {
				if !seen[symbol] {
					seen[symbol] = true
					crypto := CryptoListItem{
						Symbol:    symbol,
						AssetCode: strings.TrimSuffix(symbol, "-USD"),
						Name:      name,
					}
					cryptos = append(cryptos, crypto)
					if debug {
						fmt.Printf("Debug: Found crypto at %s: %+v\n", path, crypto)
					}
				}
			} else if ticker != "" && !strings.Contains(ticker, "-") {
				// Convert ticker to symbol format
				symbol = ticker + "-USD"
				if !seen[symbol] {
					seen[symbol] = true
					crypto := CryptoListItem{
						Symbol:    symbol,
						AssetCode: ticker,
						Name:      name,
					}
					cryptos = append(cryptos, crypto)
					if debug {
						fmt.Printf("Debug: Found crypto at %s: %+v\n", path, crypto)
					}
				}
			}
			
			// Recursively traverse
			for key, v := range val {
				newPath := path + "." + key
				if debug && (strings.Contains(key, "crypto") || strings.Contains(key, "coin") || 
					strings.Contains(key, "asset") || strings.Contains(key, "symbol") ||
					strings.Contains(key, "list") || strings.Contains(key, "item")) {
					fmt.Printf("Debug: Exploring path %s\n", newPath)
				}
				traverse(v, newPath)
			}
		case []interface{}:
			for i, item := range val {
				traverse(item, fmt.Sprintf("%s[%d]", path, i))
			}
		case string:
			// Check if this is a crypto symbol in string format
			if strings.HasSuffix(val, "-USD") && len(val) > 4 && !seen[val] {
				seen[val] = true
				cryptos = append(cryptos, CryptoListItem{
					Symbol:    val,
					AssetCode: strings.TrimSuffix(val, "-USD"),
					Name:      strings.TrimSuffix(val, "-USD"),
				})
				if debug {
					fmt.Printf("Debug: Found crypto string at %s: %s\n", path, val)
				}
			}
		}
	}
	
	traverse(data, "root")
	
	if debug {
		fmt.Printf("Debug: Found %d cryptos in JSON\n", len(cryptos))
	}
	
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