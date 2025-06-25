package models

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