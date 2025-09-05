package jobmeta

import (
	"time"

	"golang.org/x/time/rate"
)

// NewRateLimiter creates a new rate limiter based on burst and rate limit parameters
// Returns nil if burst or rateLimit is 0 (no rate limiting)
func NewRateLimiter(burst, limit int, interval time.Duration) *rate.Limiter {
	if burst > 0 && limit > 0 {
		return rate.NewLimiter(rate.Every(interval/time.Duration(limit)), burst)
	}
	return nil
}
