package ratelimit

import (
	"context"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 5, time.Second)
	if rl == nil {
		t.Fatal("expected non-nil rate limiter")
	}
	
	if rl.maxCapacity != 10 {
		t.Errorf("maxCapacity = %d, want 10", rl.maxCapacity)
	}
	if rl.refillAmount != 5 {
		t.Errorf("refillAmount = %d, want 5", rl.refillAmount)
	}
	if rl.refillInterval != time.Second {
		t.Errorf("refillInterval = %v, want %v", rl.refillInterval, time.Second)
	}
}

func TestDefaultRateLimiter(t *testing.T) {
	rl := DefaultRateLimiter()
	if rl == nil {
		t.Fatal("expected non-nil rate limiter")
	}
	
	// Check default configuration
	if rl.maxCapacity != 300 {
		t.Errorf("maxCapacity = %d, want 300", rl.maxCapacity)
	}
	if rl.refillAmount != 100 {
		t.Errorf("refillAmount = %d, want 100", rl.refillAmount)
	}
	if rl.refillInterval != time.Minute {
		t.Errorf("refillInterval = %v, want %v", rl.refillInterval, time.Minute)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	// Create a rate limiter with small capacity for testing
	rl := NewRateLimiter(3, 1, 100*time.Millisecond)
	
	// Should allow initial requests up to capacity
	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Errorf("request %d should be allowed", i+1)
		}
	}
	
	// Should deny when capacity is exhausted
	if rl.Allow() {
		t.Error("request should be denied when capacity exhausted")
	}
	
	// Wait for refill
	time.Sleep(150 * time.Millisecond)
	
	// Should allow one more request after refill
	if !rl.Allow() {
		t.Error("request should be allowed after refill")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	ctx := context.Background()
	rl := NewRateLimiter(2, 1, 50*time.Millisecond)
	
	// Exhaust initial capacity
	for i := 0; i < 2; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Errorf("Wait() error = %v", err)
		}
	}
	
	// Next wait should block but succeed after refill
	start := time.Now()
	if err := rl.Wait(ctx); err != nil {
		t.Errorf("Wait() error = %v", err)
	}
	elapsed := time.Since(start)
	
	// Should have waited approximately the refill interval
	if elapsed < 40*time.Millisecond || elapsed > 100*time.Millisecond {
		t.Errorf("Wait blocked for %v, expected ~50ms", elapsed)
	}
}

func TestRateLimiter_WaitWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	rl := NewRateLimiter(1, 1, time.Second)
	
	// Exhaust capacity
	if err := rl.Wait(ctx); err != nil {
		t.Errorf("Wait() error = %v", err)
	}
	
	// Cancel context
	cancel()
	
	// Wait should return context error
	err := rl.Wait(ctx)
	if err != context.Canceled {
		t.Errorf("Wait() error = %v, want context.Canceled", err)
	}
}

func TestRateLimiter_Reserve(t *testing.T) {
	rl := NewRateLimiter(5, 1, 100*time.Millisecond)
	
	// Make reservations
	reservations := make([]*rate.Reservation, 3)
	for i := range reservations {
		reservations[i] = rl.Reserve()
		if !reservations[i].OK() {
			t.Errorf("reservation %d should be OK", i)
		}
	}
	
	// Cancel one reservation
	reservations[1].Cancel()
	
	// Should be able to make another reservation immediately
	r := rl.Reserve()
	if !r.OK() {
		t.Error("reservation should be OK after cancellation")
	}
	if r.Delay() > 0 {
		t.Errorf("reservation delay = %v, want 0", r.Delay())
	}
}

func TestRateLimiter_Tokens(t *testing.T) {
	rl := NewRateLimiter(10, 2, 100*time.Millisecond)
	
	// Should start with max capacity
	initialTokens := rl.Tokens()
	if initialTokens < 9 || initialTokens > 10 {
		t.Errorf("initial tokens = %f, want ~10", initialTokens)
	}
	
	// Consume some tokens
	for i := 0; i < 5; i++ {
		rl.Allow()
	}
	
	// Check remaining tokens (allow for slight timing variations)
	remainingTokens := rl.Tokens()
	if remainingTokens < 3.5 || remainingTokens > 5.5 {
		t.Errorf("remaining tokens = %f, want ~5", remainingTokens)
	}
}

func TestRateLimiter_SetBurst(t *testing.T) {
	rl := NewRateLimiter(5, 1, 100*time.Millisecond)
	
	// Consume some but not all tokens
	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Errorf("initial request %d should be allowed", i)
		}
	}
	
	// Check we have some tokens left
	initialTokens := rl.Tokens()
	
	// Increase burst size
	rl.SetBurst(10)
	
	// After increasing burst, we should have more tokens
	// The exact amount depends on timing and implementation
	afterTokens := rl.Tokens()
	
	// We should have at least as many tokens as before
	if afterTokens < initialTokens {
		t.Errorf("tokens after burst increase = %f, want >= %f", afterTokens, initialTokens)
	}
	
	// Verify the new burst limit works
	rl.SetBurst(3)
	// Wait a bit to ensure we're at capacity
	time.Sleep(500 * time.Millisecond)
	
	// Should be able to make exactly 3 requests
	allowed := 0
	for i := 0; i < 5; i++ {
		if rl.Allow() {
			allowed++
		}
	}
	
	if allowed != 3 {
		t.Errorf("allowed = %d, want 3 (new burst size)", allowed)
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(100, 10, 10*time.Millisecond)
	ctx := context.Background()
	
	// Run concurrent operations
	done := make(chan bool, 3)
	
	// Goroutine 1: Wait operations
	go func() {
		for i := 0; i < 20; i++ {
			_ = rl.Wait(ctx)
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()
	
	// Goroutine 2: Allow operations
	go func() {
		for i := 0; i < 20; i++ {
			_ = rl.Allow()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()
	
	// Goroutine 3: Mixed operations
	go func() {
		for i := 0; i < 10; i++ {
			_ = rl.Tokens()
			rl.SetBurst(100 + i)
			_ = rl.Reserve()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()
	
	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for goroutines")
		}
	}
}