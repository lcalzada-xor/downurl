package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Limiter implements token bucket rate limiting
type Limiter struct {
	rate     int           // requests per period
	period   time.Duration // time period
	tokens   int           // current available tokens
	maxTokens int          // maximum tokens
	mu       sync.Mutex
	lastRefill time.Time
}

// NewLimiter creates a new rate limiter
// rate: number of requests per period
// period: time period (e.g., time.Minute)
func NewLimiter(rate int, period time.Duration) *Limiter {
	return &Limiter{
		rate:      rate,
		period:    period,
		tokens:    rate,
		maxTokens: rate,
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if allowed, waitTime := l.tryAcquire(); allowed {
			return nil
		} else if waitTime > 0 {
			select {
			case <-time.After(waitTime):
				// Continue to next iteration
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// tryAcquire attempts to acquire a token
// Returns (allowed, waitTime)
func (l *Limiter) tryAcquire() (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	if l.tokens > 0 {
		l.tokens--
		return true, 0
	}

	// Calculate wait time until next refill
	elapsed := time.Since(l.lastRefill)
	waitTime := l.period - elapsed

	return false, waitTime
}

// refill adds tokens based on time elapsed
func (l *Limiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.lastRefill)

	if elapsed >= l.period {
		// Refill all tokens
		l.tokens = l.maxTokens
		l.lastRefill = now
	}
}

// GetStatus returns current limiter status
func (l *Limiter) GetStatus() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	return fmt.Sprintf("%d/%d tokens available", l.tokens, l.maxTokens)
}

// ParseRateLimit parses rate limit string like "10/minute", "100/hour"
func ParseRateLimit(s string) (*Limiter, error) {
	// Parse format: "number/period"
	parts := []rune(s)
	slashIdx := -1
	for i, r := range parts {
		if r == '/' {
			slashIdx = i
			break
		}
	}

	if slashIdx == -1 {
		return nil, fmt.Errorf("invalid rate limit format: %s (expected number/period)", s)
	}

	rateStr := string(parts[:slashIdx])
	periodStr := string(parts[slashIdx+1:])

	var rate int
	_, err := fmt.Sscanf(rateStr, "%d", &rate)
	if err != nil {
		return nil, fmt.Errorf("invalid rate number: %s", rateStr)
	}

	var period time.Duration
	switch periodStr {
	case "second", "s":
		period = time.Second
	case "minute", "min", "m":
		period = time.Minute
	case "hour", "h":
		period = time.Hour
	default:
		// Try parsing as duration
		period, err = time.ParseDuration(periodStr)
		if err != nil {
			return nil, fmt.Errorf("invalid period: %s", periodStr)
		}
	}

	return NewLimiter(rate, period), nil
}
