package models

import "time"

type TradingPair struct {
	AssetCode       string `json:"asset_code"`
	QuoteCode       string `json:"quote_code"`
	QuoteIncrement  string `json:"quote_increment"`
	AssetIncrement  string `json:"asset_increment"`
	MaxOrderSize    string `json:"max_order_size"`
	MinOrderSize    string `json:"min_order_size"`
	Status          string `json:"status"`
	Symbol          string `json:"symbol"`
}

type TradingPairsResponse struct {
	Next     string        `json:"next"`
	Previous string        `json:"previous"`
	Results  []TradingPair `json:"results"`
}

type Holding struct {
	AccountNumber               string  `json:"account_number"`
	AssetCode                   string  `json:"asset_code"`
	TotalQuantity               float64 `json:"total_quantity"`
	QuantityAvailableForTrading float64 `json:"quantity_available_for_trading"`
}

type HoldingsResponse struct {
	Next     string    `json:"next"`
	Previous string    `json:"previous"`
	Results  []Holding `json:"results"`
}

type Execution struct {
	EffectivePrice string    `json:"effective_price"`
	Quantity       string    `json:"quantity"`
	Timestamp      time.Time `json:"timestamp"`
}

type MarketOrderConfig struct {
	AssetQuantity float64 `json:"asset_quantity,omitempty"`
	QuoteAmount   float64 `json:"quote_amount,omitempty"`
}

type LimitOrderConfig struct {
	AssetQuantity float64 `json:"asset_quantity,omitempty"`
	QuoteAmount   float64 `json:"quote_amount,omitempty"`
	LimitPrice    float64 `json:"limit_price"`
	TimeInForce   string  `json:"time_in_force"`
}

type StopLossOrderConfig struct {
	AssetQuantity float64 `json:"asset_quantity,omitempty"`
	QuoteAmount   float64 `json:"quote_amount,omitempty"`
	StopPrice     float64 `json:"stop_price"`
	TimeInForce   string  `json:"time_in_force"`
}

type StopLimitOrderConfig struct {
	AssetQuantity float64 `json:"asset_quantity,omitempty"`
	QuoteAmount   float64 `json:"quote_amount,omitempty"`
	LimitPrice    float64 `json:"limit_price"`
	StopPrice     float64 `json:"stop_price"`
	TimeInForce   string  `json:"time_in_force"`
}

type Order struct {
	ID                    string                `json:"id"`
	AccountNumber         string                `json:"account_number"`
	Symbol                string                `json:"symbol"`
	ClientOrderID         string                `json:"client_order_id"`
	Side                  string                `json:"side"`
	Executions            []Execution           `json:"executions"`
	Type                  string                `json:"type"`
	State                 string                `json:"state"`
	AveragePrice          float64               `json:"average_price"`
	FilledAssetQuantity   float64               `json:"filled_asset_quantity"`
	CreatedAt             string                `json:"created_at"`
	UpdatedAt             string                `json:"updated_at"`
	MarketOrderConfig     *MarketOrderConfig    `json:"market_order_config,omitempty"`
	LimitOrderConfig      *LimitOrderConfig     `json:"limit_order_config,omitempty"`
	StopLossOrderConfig   *StopLossOrderConfig  `json:"stop_loss_order_config,omitempty"`
	StopLimitOrderConfig  *StopLimitOrderConfig `json:"stop_limit_order_config,omitempty"`
}

type OrdersResponse struct {
	Next     string  `json:"next"`
	Previous string  `json:"previous"`
	Results  []Order `json:"results"`
}

type PlaceOrderRequest struct {
	Symbol               string                `json:"symbol"`
	ClientOrderID        string                `json:"client_order_id"`
	Side                 string                `json:"side"`
	Type                 string                `json:"type"`
	MarketOrderConfig    *MarketOrderConfig    `json:"market_order_config,omitempty"`
	LimitOrderConfig     *LimitOrderConfig     `json:"limit_order_config,omitempty"`
	StopLossOrderConfig  *StopLossOrderConfig  `json:"stop_loss_order_config,omitempty"`
	StopLimitOrderConfig *StopLimitOrderConfig `json:"stop_limit_order_config,omitempty"`
}

type OrdersFilter struct {
	CreatedAtStart *time.Time
	CreatedAtEnd   *time.Time
	UpdatedAtStart *time.Time
	UpdatedAtEnd   *time.Time
	Symbol         string
	ID             string
	Side           string
	State          string
	Type           string
	Cursor         string
	Limit          int
}