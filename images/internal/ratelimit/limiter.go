package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter manages rate limits for API keys and IP addresses using token bucket algorithm
type Limiter struct {
	limiters   map[string]*bucketInfo
	ipLimiters map[string]*bucketInfo // IP-based rate limiters
	mu         sync.RWMutex
	rates      map[string]int // plan -> requests per hour
}

// bucketInfo holds rate limiter state for a single API key
type bucketInfo struct {
	limiter     *rate.Limiter
	count       int       // Total requests in current hour window
	windowStart time.Time // When current hour window started
	lastAccess  time.Time // For cleanup of inactive limiters
}

// NewLimiter creates a new rate limiter with plan-based quotas
func NewLimiter() *Limiter {
	return &Limiter{
		limiters:   make(map[string]*bucketInfo),
		ipLimiters: make(map[string]*bucketInfo),
		rates: map[string]int{
			"test":    100,   // 100 requests/hour
			"starter": 1000,  // 1000 requests/hour
			"pro":     10000, // 10000 requests/hour
		},
	}
}

// Allow checks if a request is allowed for the given API key and plan.
// Returns: allowed, limit, remaining tokens, reset time
func (l *Limiter) Allow(apiKey string, plan string) (bool, int, int, time.Time) {
	limit := l.getRateForPlan(plan)

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get or create bucket for this API key
	bucket, exists := l.limiters[apiKey]
	if !exists {
		bucket = l.createBucket(plan)
		l.limiters[apiKey] = bucket
	}

	// Check if we need to reset the hour window
	now := time.Now()
	if now.Sub(bucket.windowStart) >= time.Hour {
		// New hour window - reset count
		bucket.count = 0
		bucket.windowStart = now
	}

	// Update last access time
	bucket.lastAccess = now

	// Check if token is available
	allowed := bucket.limiter.Allow()

	if allowed {
		bucket.count++
	}

	// Calculate remaining tokens
	remaining := limit - bucket.count
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time (start of next hour)
	resetAt := bucket.windowStart.Add(time.Hour)

	return allowed, limit, remaining, resetAt
}

// GetRequestCount returns the number of requests made in the current hour window
func (l *Limiter) GetRequestCount(apiKey string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	bucket, exists := l.limiters[apiKey]
	if !exists {
		return 0
	}

	// Check if window has expired
	if time.Since(bucket.windowStart) >= time.Hour {
		return 0
	}

	return bucket.count
}

// Cleanup removes limiters that haven't been accessed for maxAge duration
func (l *Limiter) Cleanup(maxAge time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for key, bucket := range l.limiters {
		if now.Sub(bucket.lastAccess) > maxAge {
			delete(l.limiters, key)
		}
	}
}

// createBucket creates a new token bucket for a plan
func (l *Limiter) createBucket(plan string) *bucketInfo {
	limit := l.getRateForPlan(plan)

	// Token bucket parameters:
	// - Burst = hourly limit (allows full quota in burst)
	// - Rate = limit/3600 (tokens per second for smooth refill)
	rateLimiter := rate.NewLimiter(rate.Limit(float64(limit)/3600.0), limit)

	return &bucketInfo{
		limiter:     rateLimiter,
		count:       0,
		windowStart: time.Now(),
		lastAccess:  time.Now(),
	}
}

// getRateForPlan returns the rate limit for a given plan
func (l *Limiter) getRateForPlan(plan string) int {
	if limit, ok := l.rates[plan]; ok {
		return limit
	}
	// Default to test plan quota if unknown
	return l.rates["test"]
}

// AllowIP checks if a request is allowed for the given IP address with a custom limit.
// Returns: allowed, limit, remaining tokens, reset time
func (l *Limiter) AllowIP(ipAddress string, limit int, window time.Duration) (bool, int, int, time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Get or create bucket for this IP
	bucket, exists := l.ipLimiters[ipAddress]
	if !exists {
		bucket = l.createIPBucket(limit, window)
		l.ipLimiters[ipAddress] = bucket
	}

	// Check if we need to reset the window
	now := time.Now()
	if now.Sub(bucket.windowStart) >= window {
		// New window - reset count
		bucket.count = 0
		bucket.windowStart = now
	}

	// Update last access time
	bucket.lastAccess = now

	// Check if token is available
	allowed := bucket.limiter.Allow()

	if allowed {
		bucket.count++
	}

	// Calculate remaining tokens
	remaining := limit - bucket.count
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time (start of next window)
	resetAt := bucket.windowStart.Add(window)

	return allowed, limit, remaining, resetAt
}

// createIPBucket creates a new token bucket for IP-based rate limiting
func (l *Limiter) createIPBucket(limit int, window time.Duration) *bucketInfo {
	// Token bucket parameters:
	// - Burst = limit (allows full quota in burst)
	// - Rate = limit/window.Seconds() (tokens per second for smooth refill)
	rateLimiter := rate.NewLimiter(rate.Limit(float64(limit)/window.Seconds()), limit)

	return &bucketInfo{
		limiter:     rateLimiter,
		count:       0,
		windowStart: time.Now(),
		lastAccess:  time.Now(),
	}
}
