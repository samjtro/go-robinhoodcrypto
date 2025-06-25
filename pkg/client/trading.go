package client

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/samjtro/go-robinhoodcrypto/pkg/models"
)

// GetTradingPairs fetches the list of available trading pairs
func (s *TradingService) GetTradingPairs(ctx context.Context, symbols ...string) (*models.TradingPairsResponse, error) {
	query := url.Values{}
	for _, symbol := range symbols {
		query.Add("symbol", strings.ToUpper(symbol))
	}

	var result models.TradingPairsResponse
	err := s.client.do(ctx, "GET", "/api/v1/crypto/trading/trading_pairs/", query, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetHoldings fetches the list of crypto holdings
func (s *TradingService) GetHoldings(ctx context.Context, assetCodes ...string) (*models.HoldingsResponse, error) {
	query := url.Values{}
	for _, code := range assetCodes {
		query.Add("asset_code", strings.ToUpper(code))
	}

	var result models.HoldingsResponse
	err := s.client.do(ctx, "GET", "/api/v1/crypto/trading/holdings/", query, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOrder fetches a specific order by ID
func (s *TradingService) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	path := fmt.Sprintf("/api/v1/crypto/trading/orders/%s/", orderID)
	
	var result models.Order
	err := s.client.do(ctx, "GET", path, nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOrders fetches a list of orders with optional filters
func (s *TradingService) GetOrders(ctx context.Context, filter *models.OrdersFilter) (*models.OrdersResponse, error) {
	query := url.Values{}
	
	if filter != nil {
		if filter.CreatedAtStart != nil {
			query.Set("created_at_start", filter.CreatedAtStart.Format("2006-01-02T15:04:05Z"))
		}
		if filter.CreatedAtEnd != nil {
			query.Set("created_at_end", filter.CreatedAtEnd.Format("2006-01-02T15:04:05Z"))
		}
		if filter.UpdatedAtStart != nil {
			query.Set("updated_at_start", filter.UpdatedAtStart.Format("2006-01-02T15:04:05Z"))
		}
		if filter.UpdatedAtEnd != nil {
			query.Set("updated_at_end", filter.UpdatedAtEnd.Format("2006-01-02T15:04:05Z"))
		}
		if filter.Symbol != "" {
			query.Set("symbol", strings.ToUpper(filter.Symbol))
		}
		if filter.ID != "" {
			query.Set("id", filter.ID)
		}
		if filter.Side != "" {
			query.Set("side", filter.Side)
		}
		if filter.State != "" {
			query.Set("state", filter.State)
		}
		if filter.Type != "" {
			query.Set("type", filter.Type)
		}
		if filter.Cursor != "" {
			query.Set("cursor", filter.Cursor)
		}
		if filter.Limit > 0 {
			query.Set("limit", fmt.Sprintf("%d", filter.Limit))
		}
	}

	var result models.OrdersResponse
	err := s.client.do(ctx, "GET", "/api/v1/crypto/trading/orders/", query, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// PlaceOrder places a new crypto order
func (s *TradingService) PlaceOrder(ctx context.Context, req *models.PlaceOrderRequest) (*models.Order, error) {
	// Validate request
	if err := s.validateOrderRequest(req); err != nil {
		return nil, err
	}

	// Ensure symbol is uppercase
	req.Symbol = strings.ToUpper(req.Symbol)

	// Create the request body with the correct order config field name
	body := map[string]interface{}{
		"symbol":          req.Symbol,
		"client_order_id": req.ClientOrderID,
		"side":            req.Side,
		"type":            req.Type,
	}

	// Add the appropriate order config based on type
	switch req.Type {
	case "market":
		if req.MarketOrderConfig != nil {
			body["market_order_config"] = req.MarketOrderConfig
		}
	case "limit":
		if req.LimitOrderConfig != nil {
			body["limit_order_config"] = req.LimitOrderConfig
		}
	case "stop_loss":
		if req.StopLossOrderConfig != nil {
			body["stop_loss_order_config"] = req.StopLossOrderConfig
		}
	case "stop_limit":
		if req.StopLimitOrderConfig != nil {
			body["stop_limit_order_config"] = req.StopLimitOrderConfig
		}
	}

	var result models.Order
	err := s.client.do(ctx, "POST", "/api/v1/crypto/trading/orders/", nil, body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CancelOrder cancels an open crypto order
func (s *TradingService) CancelOrder(ctx context.Context, orderID string) error {
	path := fmt.Sprintf("/api/v1/crypto/trading/orders/%s/cancel/", orderID)
	return s.client.do(ctx, "POST", path, nil, nil, nil)
}

// validateOrderRequest validates the order request parameters
func (s *TradingService) validateOrderRequest(req *models.PlaceOrderRequest) error {
	if req.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if req.ClientOrderID == "" {
		return fmt.Errorf("client_order_id is required")
	}
	if req.Side != "buy" && req.Side != "sell" {
		return fmt.Errorf("invalid side: must be 'buy' or 'sell'")
	}
	
	// Validate order type and corresponding config
	switch req.Type {
	case "market":
		if req.MarketOrderConfig == nil {
			return fmt.Errorf("market_order_config is required for market orders")
		}
		if req.MarketOrderConfig.AssetQuantity == 0 && req.MarketOrderConfig.QuoteAmount == 0 {
			return fmt.Errorf("either asset_quantity or quote_amount must be specified")
		}
		if req.MarketOrderConfig.AssetQuantity != 0 && req.MarketOrderConfig.QuoteAmount != 0 {
			return fmt.Errorf("only one of asset_quantity or quote_amount can be specified")
		}
	case "limit":
		if req.LimitOrderConfig == nil {
			return fmt.Errorf("limit_order_config is required for limit orders")
		}
		if req.LimitOrderConfig.LimitPrice <= 0 {
			return fmt.Errorf("limit_price must be greater than 0")
		}
		if req.LimitOrderConfig.AssetQuantity == 0 && req.LimitOrderConfig.QuoteAmount == 0 {
			return fmt.Errorf("either asset_quantity or quote_amount must be specified")
		}
		if req.LimitOrderConfig.AssetQuantity != 0 && req.LimitOrderConfig.QuoteAmount != 0 {
			return fmt.Errorf("only one of asset_quantity or quote_amount can be specified")
		}
		if req.LimitOrderConfig.TimeInForce == "" {
			req.LimitOrderConfig.TimeInForce = "gtc"
		}
	case "stop_loss":
		if req.StopLossOrderConfig == nil {
			return fmt.Errorf("stop_loss_order_config is required for stop loss orders")
		}
		if req.StopLossOrderConfig.StopPrice <= 0 {
			return fmt.Errorf("stop_price must be greater than 0")
		}
		if req.StopLossOrderConfig.AssetQuantity == 0 && req.StopLossOrderConfig.QuoteAmount == 0 {
			return fmt.Errorf("either asset_quantity or quote_amount must be specified")
		}
		if req.StopLossOrderConfig.AssetQuantity != 0 && req.StopLossOrderConfig.QuoteAmount != 0 {
			return fmt.Errorf("only one of asset_quantity or quote_amount can be specified")
		}
		if req.StopLossOrderConfig.TimeInForce == "" {
			req.StopLossOrderConfig.TimeInForce = "gtc"
		}
	case "stop_limit":
		if req.StopLimitOrderConfig == nil {
			return fmt.Errorf("stop_limit_order_config is required for stop limit orders")
		}
		if req.StopLimitOrderConfig.StopPrice <= 0 {
			return fmt.Errorf("stop_price must be greater than 0")
		}
		if req.StopLimitOrderConfig.LimitPrice <= 0 {
			return fmt.Errorf("limit_price must be greater than 0")
		}
		if req.StopLimitOrderConfig.AssetQuantity == 0 && req.StopLimitOrderConfig.QuoteAmount == 0 {
			return fmt.Errorf("either asset_quantity or quote_amount must be specified")
		}
		if req.StopLimitOrderConfig.AssetQuantity != 0 && req.StopLimitOrderConfig.QuoteAmount != 0 {
			return fmt.Errorf("only one of asset_quantity or quote_amount can be specified")
		}
		if req.StopLimitOrderConfig.TimeInForce == "" {
			req.StopLimitOrderConfig.TimeInForce = "gtc"
		}
	default:
		return fmt.Errorf("invalid order type: must be 'market', 'limit', 'stop_loss', or 'stop_limit'")
	}

	return nil
}