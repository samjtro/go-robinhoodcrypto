package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAccountDetails_JSONMarshaling(t *testing.T) {
	original := &AccountDetails{
		AccountNumber:        "ACC123456",
		Status:              "active",
		BuyingPower:         "10000.50",
		BuyingPowerCurrency: "USD",
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal back
	var decoded AccountDetails
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Compare
	if decoded != *original {
		t.Errorf("decoded = %+v, want %+v", decoded, *original)
	}
}

func TestOrder_JSONMarshaling(t *testing.T) {
	timestamp, _ := time.Parse(time.RFC3339, "2023-10-31T20:57:50Z")
	
	original := &Order{
		ID:                  "497f6eca-6276-4993-bfeb-53cbbbba6f08",
		AccountNumber:       "ACC123456",
		Symbol:              "BTC-USD",
		ClientOrderID:       "11299b2b-61e3-43e7-b9f7-dee77210bb29",
		Side:                "buy",
		Type:                "limit",
		State:               "open",
		AveragePrice:        45000.50,
		FilledAssetQuantity: 0.1,
		CreatedAt:           "2023-10-31T20:57:50Z",
		UpdatedAt:           "2023-10-31T20:58:00Z",
		Executions: []Execution{
			{
				EffectivePrice: "45000.00",
				Quantity:       "0.05",
				Timestamp:      timestamp,
			},
		},
		LimitOrderConfig: &LimitOrderConfig{
			AssetQuantity: 0.2,
			LimitPrice:    45000,
			TimeInForce:   "gtc",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Unmarshal back
	var decoded Order
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Basic field comparison
	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Symbol != original.Symbol {
		t.Errorf("Symbol = %q, want %q", decoded.Symbol, original.Symbol)
	}
	if decoded.Side != original.Side {
		t.Errorf("Side = %q, want %q", decoded.Side, original.Side)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, original.Type)
	}
	if decoded.State != original.State {
		t.Errorf("State = %q, want %q", decoded.State, original.State)
	}

	// Check nested structures
	if len(decoded.Executions) != len(original.Executions) {
		t.Errorf("len(Executions) = %d, want %d", len(decoded.Executions), len(original.Executions))
	}

	if decoded.LimitOrderConfig == nil {
		t.Fatal("LimitOrderConfig is nil")
	}
	if decoded.LimitOrderConfig.LimitPrice != original.LimitOrderConfig.LimitPrice {
		t.Errorf("LimitOrderConfig.LimitPrice = %f, want %f", 
			decoded.LimitOrderConfig.LimitPrice, original.LimitOrderConfig.LimitPrice)
	}
}

func TestPlaceOrderRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     PlaceOrderRequest
		wantErr bool
	}{
		{
			name: "valid market order with asset quantity",
			req: PlaceOrderRequest{
				Symbol:        "BTC-USD",
				ClientOrderID: "123e4567-e89b-12d3-a456-426614174000",
				Side:          "buy",
				Type:          "market",
				MarketOrderConfig: &MarketOrderConfig{
					AssetQuantity: 0.1,
				},
			},
			wantErr: false,
		},
		{
			name: "valid limit order",
			req: PlaceOrderRequest{
				Symbol:        "ETH-USD",
				ClientOrderID: "123e4567-e89b-12d3-a456-426614174001",
				Side:          "sell",
				Type:          "limit",
				LimitOrderConfig: &LimitOrderConfig{
					AssetQuantity: 1.5,
					LimitPrice:    2500.00,
					TimeInForce:   "gtc",
				},
			},
			wantErr: false,
		},
		{
			name: "valid stop loss order",
			req: PlaceOrderRequest{
				Symbol:        "BTC-USD",
				ClientOrderID: "123e4567-e89b-12d3-a456-426614174002",
				Side:          "sell",
				Type:          "stop_loss",
				StopLossOrderConfig: &StopLossOrderConfig{
					AssetQuantity: 0.5,
					StopPrice:     40000.00,
					TimeInForce:   "gtc",
				},
			},
			wantErr: false,
		},
		{
			name: "valid stop limit order",
			req: PlaceOrderRequest{
				Symbol:        "BTC-USD",
				ClientOrderID: "123e4567-e89b-12d3-a456-426614174003",
				Side:          "buy",
				Type:          "stop_limit",
				StopLimitOrderConfig: &StopLimitOrderConfig{
					AssetQuantity: 0.25,
					StopPrice:     42000.00,
					LimitPrice:    42500.00,
					TimeInForce:   "ioc",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.req)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}

			// Verify the JSON includes the correct fields
			var decoded map[string]interface{}
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
				return
			}

			// Check required fields are present
			requiredFields := []string{"symbol", "client_order_id", "side", "type"}
			for _, field := range requiredFields {
				if _, ok := decoded[field]; !ok {
					t.Errorf("missing required field %q in JSON", field)
				}
			}

			// Check that only the appropriate order config is included
			configFields := []string{
				"market_order_config",
				"limit_order_config", 
				"stop_loss_order_config",
				"stop_limit_order_config",
			}
			
			expectedConfig := tt.req.Type + "_order_config"
			foundExpected := false
			
			for _, field := range configFields {
				_, exists := decoded[field]
				if field == expectedConfig {
					if !exists {
						t.Errorf("expected %q to be present in JSON", field)
					}
					foundExpected = true
				} else {
					if exists && decoded[field] != nil {
						t.Errorf("unexpected %q in JSON for %s order", field, tt.req.Type)
					}
				}
			}
			
			if !foundExpected && tt.req.Type != "" {
				t.Errorf("no order config found for %s order type", tt.req.Type)
			}
		})
	}
}

func TestTradingPair_JSONMarshaling(t *testing.T) {
	original := &TradingPair{
		AssetCode:      "BTC",
		QuoteCode:      "USD",
		QuoteIncrement: "0.01",
		AssetIncrement: "0.00000001",
		MaxOrderSize:   "100",
		MinOrderSize:   "0.0001",
		Status:         "tradable",
		Symbol:         "BTC-USD",
	}

	// Test round-trip JSON marshaling
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded TradingPair
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded != *original {
		t.Errorf("decoded = %+v, want %+v", decoded, *original)
	}
}

func TestPaginatedResponses(t *testing.T) {
	// Test TradingPairsResponse
	tradingResp := &TradingPairsResponse{
		Next:     "https://api.example.com/next?cursor=abc123",
		Previous: "https://api.example.com/prev?cursor=xyz789",
		Results: []TradingPair{
			{Symbol: "BTC-USD", Status: "tradable"},
			{Symbol: "ETH-USD", Status: "tradable"},
		},
	}

	data, err := json.Marshal(tradingResp)
	if err != nil {
		t.Fatalf("json.Marshal(TradingPairsResponse) error = %v", err)
	}

	var decodedTrading TradingPairsResponse
	if err := json.Unmarshal(data, &decodedTrading); err != nil {
		t.Fatalf("json.Unmarshal(TradingPairsResponse) error = %v", err)
	}

	if decodedTrading.Next != tradingResp.Next {
		t.Errorf("Next = %q, want %q", decodedTrading.Next, tradingResp.Next)
	}
	if len(decodedTrading.Results) != len(tradingResp.Results) {
		t.Errorf("len(Results) = %d, want %d", len(decodedTrading.Results), len(tradingResp.Results))
	}
}