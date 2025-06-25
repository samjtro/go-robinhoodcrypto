package client

import (
	"context"
	"net/url"
	"strings"

	"github.com/samjtro/go-robinhoodcrypto/pkg/models"
)

// PaginationOptions provides pagination control
type PaginationOptions struct {
	Cursor string
	Limit  int
}

// Paginator helps iterate through paginated results
type Paginator[T any] struct {
	client   *Client
	nextURL  string
	prevURL  string
	fetcher  func(ctx context.Context, cursor string) (*PaginatedResponse[T], error)
}

// PaginatedResponse is a generic wrapper for paginated responses
type PaginatedResponse[T any] struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

// HasNext returns true if there are more pages
func (p *Paginator[T]) HasNext() bool {
	return p.nextURL != ""
}

// HasPrevious returns true if there are previous pages
func (p *Paginator[T]) HasPrevious() bool {
	return p.prevURL != ""
}

// Next fetches the next page of results
func (p *Paginator[T]) Next(ctx context.Context) ([]T, error) {
	if !p.HasNext() {
		return nil, nil
	}

	cursor := extractCursor(p.nextURL)
	resp, err := p.fetcher(ctx, cursor)
	if err != nil {
		return nil, err
	}

	p.nextURL = resp.Next
	p.prevURL = resp.Previous
	return resp.Results, nil
}

// Previous fetches the previous page of results
func (p *Paginator[T]) Previous(ctx context.Context) ([]T, error) {
	if !p.HasPrevious() {
		return nil, nil
	}

	cursor := extractCursor(p.prevURL)
	resp, err := p.fetcher(ctx, cursor)
	if err != nil {
		return nil, err
	}

	p.nextURL = resp.Next
	p.prevURL = resp.Previous
	return resp.Results, nil
}

// extractCursor extracts the cursor parameter from a URL
func extractCursor(urlStr string) string {
	if urlStr == "" {
		return ""
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	return u.Query().Get("cursor")
}

// GetAllPages fetches all pages of results
func (p *Paginator[T]) GetAllPages(ctx context.Context) ([]T, error) {
	var allResults []T
	
	// Get initial results
	if p.nextURL != "" {
		results, err := p.Next(ctx)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	// Get remaining pages
	for p.HasNext() {
		results, err := p.Next(ctx)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// Helper methods for creating paginators

// NewTradingPairsPaginator creates a paginator for trading pairs
func (s *TradingService) NewTradingPairsPaginator(symbols ...string) *Paginator[models.TradingPair] {
	return &Paginator[models.TradingPair]{
		client: s.client,
		fetcher: func(ctx context.Context, cursor string) (*PaginatedResponse[models.TradingPair], error) {
			query := url.Values{}
			for _, symbol := range symbols {
				query.Add("symbol", strings.ToUpper(symbol))
			}
			if cursor != "" {
				query.Set("cursor", cursor)
			}

			var result models.TradingPairsResponse
			err := s.client.do(ctx, "GET", "/api/v1/crypto/trading/trading_pairs/", query, nil, &result)
			if err != nil {
				return nil, err
			}

			return &PaginatedResponse[models.TradingPair]{
				Next:     result.Next,
				Previous: result.Previous,
				Results:  result.Results,
			}, nil
		},
	}
}

// NewHoldingsPaginator creates a paginator for holdings
func (s *TradingService) NewHoldingsPaginator(assetCodes ...string) *Paginator[models.Holding] {
	return &Paginator[models.Holding]{
		client: s.client,
		fetcher: func(ctx context.Context, cursor string) (*PaginatedResponse[models.Holding], error) {
			query := url.Values{}
			for _, code := range assetCodes {
				query.Add("asset_code", strings.ToUpper(code))
			}
			if cursor != "" {
				query.Set("cursor", cursor)
			}

			var result models.HoldingsResponse
			err := s.client.do(ctx, "GET", "/api/v1/crypto/trading/holdings/", query, nil, &result)
			if err != nil {
				return nil, err
			}

			return &PaginatedResponse[models.Holding]{
				Next:     result.Next,
				Previous: result.Previous,
				Results:  result.Results,
			}, nil
		},
	}
}

// NewOrdersPaginator creates a paginator for orders
func (s *TradingService) NewOrdersPaginator(filter *models.OrdersFilter) *Paginator[models.Order] {
	return &Paginator[models.Order]{
		client: s.client,
		fetcher: func(ctx context.Context, cursor string) (*PaginatedResponse[models.Order], error) {
			// Copy filter to avoid modifying the original
			localFilter := &models.OrdersFilter{}
			if filter != nil {
				*localFilter = *filter
			}
			localFilter.Cursor = cursor

			resp, err := s.GetOrders(ctx, localFilter)
			if err != nil {
				return nil, err
			}

			return &PaginatedResponse[models.Order]{
				Next:     resp.Next,
				Previous: resp.Previous,
				Results:  resp.Results,
			}, nil
		},
	}
}