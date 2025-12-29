package main

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	rate       float64 // tokens per second
	capacity   int     // maximum tokens
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		capacity:   capacity,
		tokens:     float64(capacity),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		// Try to acquire token
		if rl.tryAcquire() {
			return nil
		}

		// Check context
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Calculate wait time and sleep
		waitTime := rl.calculateWaitTime()
		if waitTime > 0 {
			select {
			case <-time.After(waitTime):
				// Try again
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// tryAcquire attempts to acquire a token without blocking
func (rl *RateLimiter) tryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// refill adds tokens based on elapsed time
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()

	if elapsed > 0 {
		tokensToAdd := elapsed * rl.rate
		rl.tokens += tokensToAdd

		if rl.tokens > float64(rl.capacity) {
			rl.tokens = float64(rl.capacity)
		}

		rl.lastRefill = now
	}
}

// calculateWaitTime calculates how long to wait for next token
func (rl *RateLimiter) calculateWaitTime() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens >= 1.0 {
		return 0
	}

	tokensNeeded := 1.0 - rl.tokens
	waitSeconds := tokensNeeded / rl.rate

	// Add small buffer to avoid tight loops
	return time.Duration(waitSeconds*1.1) * time.Second
}

// RateLimiterRegistry manages rate limiters for missions
type RateLimiterRegistry struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
}

// NewRateLimiterRegistry creates a new registry
func NewRateLimiterRegistry() *RateLimiterRegistry {
	return &RateLimiterRegistry{
		limiters: make(map[string]*RateLimiter),
	}
}

// Get gets or creates a rate limiter for a mission
func (r *RateLimiterRegistry) Get(missionID string, rate float64) *RateLimiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	if limiter, exists := r.limiters[missionID]; exists {
		return limiter
	}

	// Create new limiter with capacity = rate (burst of 1 second)
	capacity := int(rate) + 1
	if capacity < 1 {
		capacity = 1
	}

	limiter := NewRateLimiter(rate, capacity)
	r.limiters[missionID] = limiter

	return limiter
}

// Remove removes a rate limiter
func (r *RateLimiterRegistry) Remove(missionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.limiters, missionID)
}
