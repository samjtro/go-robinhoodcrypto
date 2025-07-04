package ratelimit

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter implements token bucket rate limiting for the API
type RateLimiter struct {
	limiter *rate.Limiter
	mu      sync.Mutex
	
	// Configuration
	maxCapacity   int
	refillAmount  int
	refillInterval time.Duration
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(maxCapacity, refillAmount int, refillInterval time.Duration) *RateLimiter {
	// Create a rate limiter with the specified refill rate
	refillRate := rate.Every(refillInterval / time.Duration(refillAmount))
	limiter := rate.NewLimiter(refillRate, maxCapacity)
	
	return &RateLimiter{
		limiter:        limiter,
		maxCapacity:    maxCapacity,
		refillAmount:   refillAmount,
		refillInterval: refillInterval,
	}
}

// DefaultRateLimiter creates a rate limiter with Robinhood's default limits
// 100 requests per minute normally, 300 in bursts
func DefaultRateLimiter() *RateLimiter {
	// 100 requests per minute = ~1.67 requests per second
	// Burst capacity of 300
	return NewRateLimiter(300, 100, time.Minute)
}

// Wait blocks until a token is available or the context is cancelled
func (r *RateLimiter) Wait(ctx context.Context) error {
	return r.limiter.Wait(ctx)
}

// Allow reports whether an event may happen now
func (r *RateLimiter) Allow() bool {
	return r.limiter.Allow()
}

// Reserve returns a Reservation that can be used to wait for or cancel
func (r *RateLimiter) Reserve() *rate.Reservation {
	return r.limiter.Reserve()
}

// Tokens returns the current number of available tokens
func (r *RateLimiter) Tokens() float64 {
	return r.limiter.Tokens()
}

// SetBurst updates the burst size (max capacity)
func (r *RateLimiter) SetBurst(burst int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.limiter.SetBurst(burst)
	r.maxCapacity = burst
}

// SetLimit updates the refill rate
func (r *RateLimiter) SetLimit(tokensPerSecond rate.Limit) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.limiter.SetLimit(tokensPerSecond)
}