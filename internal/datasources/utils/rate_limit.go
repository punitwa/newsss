// Package utils provides rate limiting utilities for data sources.
package utils

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting functionality.
type RateLimiter struct {
	limiter *rate.Limiter
	logger  zerolog.Logger
	mu      sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with the specified rate.
func NewRateLimiter(rateLimit float64, burst int, logger zerolog.Logger) *RateLimiter {
	if rateLimit <= 0 {
		rateLimit = 1.0 // Default to 1 request per second
	}

	if burst <= 0 {
		burst = 1 // Default burst of 1
	}

	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rateLimit), burst),
		logger:  logger.With().Str("component", "rate_limiter").Logger(),
	}
}

// Wait blocks until the rate limiter allows the operation.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.RLock()
	limiter := rl.limiter
	rl.mu.RUnlock()

	start := time.Now()
	err := limiter.Wait(ctx)
	duration := time.Since(start)

	if err != nil {
		rl.logger.Error().
			Err(err).
			Dur("wait_duration", duration).
			Msg("Rate limiter wait failed")
		return err
	}

	if duration > time.Millisecond {
		rl.logger.Debug().
			Dur("wait_duration", duration).
			Msg("Rate limiter applied delay")
	}

	return nil
}

// Allow checks if an operation is allowed without blocking.
func (rl *RateLimiter) Allow() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	allowed := rl.limiter.Allow()

	rl.logger.Debug().
		Bool("allowed", allowed).
		Msg("Rate limiter check")

	return allowed
}

// SetLimit updates the rate limit.
func (rl *RateLimiter) SetLimit(limit float64) {
	if limit <= 0 {
		limit = 1.0
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limiter.SetLimit(rate.Limit(limit))

	rl.logger.Info().
		Float64("new_limit", limit).
		Msg("Rate limit updated")
}

// GetLimit returns the current rate limit.
func (rl *RateLimiter) GetLimit() float64 {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return float64(rl.limiter.Limit())
}

// SetBurst updates the burst size.
func (rl *RateLimiter) SetBurst(burst int) {
	if burst <= 0 {
		burst = 1
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limiter.SetBurst(burst)

	rl.logger.Info().
		Int("new_burst", burst).
		Msg("Rate limiter burst updated")
}

// GetBurst returns the current burst size.
func (rl *RateLimiter) GetBurst() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.limiter.Burst()
}

// TokenBucket provides a token bucket rate limiter implementation.
type TokenBucket struct {
	tokens     int
	capacity   int
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
	logger     zerolog.Logger
}

// NewTokenBucket creates a new token bucket rate limiter.
func NewTokenBucket(capacity int, refillRate float64, logger zerolog.Logger) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
		logger:     logger.With().Str("component", "token_bucket").Logger(),
	}
}

// TryConsume attempts to consume a token from the bucket.
func (tb *TokenBucket) TryConsume() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens > 0 {
		tb.tokens--
		tb.logger.Debug().
			Int("remaining_tokens", tb.tokens).
			Msg("Token consumed")
		return true
	}

	tb.logger.Debug().
		Int("tokens", tb.tokens).
		Msg("No tokens available")
	return false
}

// WaitForToken waits until a token is available or context is cancelled.
func (tb *TokenBucket) WaitForToken(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if tb.TryConsume() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Continue waiting
		}
	}
}

// refill adds tokens to the bucket based on elapsed time.
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	tokensToAdd := int(elapsed * tb.refillRate)
	if tokensToAdd > 0 {
		tb.tokens = minInt(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now

		tb.logger.Debug().
			Int("tokens_added", tokensToAdd).
			Int("current_tokens", tb.tokens).
			Msg("Tokens refilled")
	}
}

// GetTokens returns the current number of tokens.
func (tb *TokenBucket) GetTokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	return tb.tokens
}

// AdaptiveRateLimiter adjusts rate limits based on success/failure rates.
type AdaptiveRateLimiter struct {
	baseLimiter      *RateLimiter
	currentLimit     float64
	baseLimit        float64
	successCount     int64
	failureCount     int64
	adjustmentPeriod time.Duration
	lastAdjustment   time.Time
	mu               sync.RWMutex
	logger           zerolog.Logger
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter.
func NewAdaptiveRateLimiter(baseLimit float64, logger zerolog.Logger) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		baseLimiter:      NewRateLimiter(baseLimit, 1, logger),
		currentLimit:     baseLimit,
		baseLimit:        baseLimit,
		adjustmentPeriod: 1 * time.Minute,
		lastAdjustment:   time.Now(),
		logger:           logger.With().Str("component", "adaptive_rate_limiter").Logger(),
	}
}

// Wait blocks until the rate limiter allows the operation.
func (arl *AdaptiveRateLimiter) Wait(ctx context.Context) error {
	return arl.baseLimiter.Wait(ctx)
}

// Allow checks if an operation is allowed without blocking.
func (arl *AdaptiveRateLimiter) Allow() bool {
	return arl.baseLimiter.Allow()
}

// RecordSuccess records a successful operation.
func (arl *AdaptiveRateLimiter) RecordSuccess() {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	arl.successCount++
	arl.maybeAdjustLimit()
}

// RecordFailure records a failed operation.
func (arl *AdaptiveRateLimiter) RecordFailure() {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	arl.failureCount++
	arl.maybeAdjustLimit()
}

// maybeAdjustLimit adjusts the rate limit based on success/failure ratio.
func (arl *AdaptiveRateLimiter) maybeAdjustLimit() {
	if time.Since(arl.lastAdjustment) < arl.adjustmentPeriod {
		return
	}

	total := arl.successCount + arl.failureCount
	if total < 10 { // Need minimum sample size
		return
	}

	successRate := float64(arl.successCount) / float64(total)

	var newLimit float64
	switch {
	case successRate > 0.95: // Very high success rate, increase limit
		newLimit = min(arl.currentLimit*1.2, arl.baseLimit*2)
	case successRate > 0.8: // Good success rate, slightly increase
		newLimit = min(arl.currentLimit*1.1, arl.baseLimit*1.5)
	case successRate < 0.5: // Poor success rate, decrease significantly
		newLimit = max(arl.currentLimit*0.5, arl.baseLimit*0.1)
	case successRate < 0.7: // Moderate success rate, decrease slightly
		newLimit = max(arl.currentLimit*0.8, arl.baseLimit*0.5)
	default:
		newLimit = arl.currentLimit // No change
	}

	if newLimit != arl.currentLimit {
		arl.currentLimit = newLimit
		arl.baseLimiter.SetLimit(newLimit)

		arl.logger.Info().
			Float64("old_limit", arl.currentLimit).
			Float64("new_limit", newLimit).
			Float64("success_rate", successRate).
			Int64("success_count", arl.successCount).
			Int64("failure_count", arl.failureCount).
			Msg("Rate limit adjusted")
	}

	// Reset counters
	arl.successCount = 0
	arl.failureCount = 0
	arl.lastAdjustment = time.Now()
}

// SetLimit updates the base rate limit.
func (arl *AdaptiveRateLimiter) SetLimit(limit float64) {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	arl.baseLimit = limit
	arl.currentLimit = limit
	arl.baseLimiter.SetLimit(limit)
}

// GetLimit returns the current rate limit.
func (arl *AdaptiveRateLimiter) GetLimit() float64 {
	arl.mu.RLock()
	defer arl.mu.RUnlock()

	return arl.currentLimit
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
