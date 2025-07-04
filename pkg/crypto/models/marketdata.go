package models

import (
	"encoding/json"
	"strconv"
)

type BestBidAskResult struct {
	Symbol                    string  `json:"symbol"`
	Price                     float64 `json:"price"`
	BidInclusiveOfSellSpread  float64 `json:"bid_inclusive_of_sell_spread"`
	SellSpread                float64 `json:"sell_spread"`
	AskInclusiveOfBuySpread   float64 `json:"ask_inclusive_of_buy_spread"`
	BuySpread                 float64 `json:"buy_spread"`
	Timestamp                 string  `json:"timestamp"`
}

type BestBidAskResponse struct {
	Results []BestBidAskResult `json:"results"`
}

type EstimatedPriceResult struct {
	Symbol                    string  `json:"symbol"`
	Side                      string  `json:"side"`
	Price                     float64 `json:"price"`
	Quantity                  float64 `json:"quantity"`
	BidInclusiveOfSellSpread  float64 `json:"bid_inclusive_of_sell_spread"`
	SellSpread                float64 `json:"sell_spread"`
	AskInclusiveOfBuySpread   float64 `json:"ask_inclusive_of_buy_spread"`
	BuySpread                 float64 `json:"buy_spread"`
	Timestamp                 string  `json:"timestamp"`
}

type EstimatedPriceResponse struct {
	Results []EstimatedPriceResult `json:"results"`
}

// UnmarshalJSON custom unmarshaler for BestBidAskResult to handle string numbers
func (b *BestBidAskResult) UnmarshalJSON(data []byte) error {
	type Alias BestBidAskResult
	aux := &struct {
		Price                    string `json:"price"`
		BidInclusiveOfSellSpread string `json:"bid_inclusive_of_sell_spread"`
		SellSpread               string `json:"sell_spread"`
		AskInclusiveOfBuySpread  string `json:"ask_inclusive_of_buy_spread"`
		BuySpread                string `json:"buy_spread"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert string numbers to float64
	if aux.Price != "" {
		if v, err := strconv.ParseFloat(aux.Price, 64); err == nil {
			b.Price = v
		}
	}
	if aux.BidInclusiveOfSellSpread != "" {
		if v, err := strconv.ParseFloat(aux.BidInclusiveOfSellSpread, 64); err == nil {
			b.BidInclusiveOfSellSpread = v
		}
	}
	if aux.SellSpread != "" {
		if v, err := strconv.ParseFloat(aux.SellSpread, 64); err == nil {
			b.SellSpread = v
		}
	}
	if aux.AskInclusiveOfBuySpread != "" {
		if v, err := strconv.ParseFloat(aux.AskInclusiveOfBuySpread, 64); err == nil {
			b.AskInclusiveOfBuySpread = v
		}
	}
	if aux.BuySpread != "" {
		if v, err := strconv.ParseFloat(aux.BuySpread, 64); err == nil {
			b.BuySpread = v
		}
	}

	return nil
}

// UnmarshalJSON custom unmarshaler for EstimatedPriceResult to handle string numbers
func (e *EstimatedPriceResult) UnmarshalJSON(data []byte) error {
	type Alias EstimatedPriceResult
	aux := &struct {
		Price                    string `json:"price"`
		Quantity                 string `json:"quantity"`
		BidInclusiveOfSellSpread string `json:"bid_inclusive_of_sell_spread"`
		SellSpread               string `json:"sell_spread"`
		AskInclusiveOfBuySpread  string `json:"ask_inclusive_of_buy_spread"`
		BuySpread                string `json:"buy_spread"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert string numbers to float64
	if aux.Price != "" {
		if v, err := strconv.ParseFloat(aux.Price, 64); err == nil {
			e.Price = v
		}
	}
	if aux.Quantity != "" {
		if v, err := strconv.ParseFloat(aux.Quantity, 64); err == nil {
			e.Quantity = v
		}
	}
	if aux.BidInclusiveOfSellSpread != "" {
		if v, err := strconv.ParseFloat(aux.BidInclusiveOfSellSpread, 64); err == nil {
			e.BidInclusiveOfSellSpread = v
		}
	}
	if aux.SellSpread != "" {
		if v, err := strconv.ParseFloat(aux.SellSpread, 64); err == nil {
			e.SellSpread = v
		}
	}
	if aux.AskInclusiveOfBuySpread != "" {
		if v, err := strconv.ParseFloat(aux.AskInclusiveOfBuySpread, 64); err == nil {
			e.AskInclusiveOfBuySpread = v
		}
	}
	if aux.BuySpread != "" {
		if v, err := strconv.ParseFloat(aux.BuySpread, 64); err == nil {
			e.BuySpread = v
		}
	}

	return nil
}