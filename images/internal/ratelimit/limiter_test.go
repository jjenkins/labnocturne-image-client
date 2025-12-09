package ratelimit

import (
	"testing"
	"time"
)

func TestAllowIP(t *testing.T) {
	limiter := NewLimiter()

	t.Run("Allow requests within limit", func(t *testing.T) {
		ip := "192.168.1.100"
		limit := 5
		window := time.Hour

		// First 5 requests should be allowed
		for i := 0; i < 5; i++ {
			allowed, _, remaining, _ := limiter.AllowIP(ip, limit, window)
			if !allowed {
				t.Errorf("Request %d should be allowed", i+1)
			}
			expectedRemaining := limit - (i + 1)
			if remaining != expectedRemaining {
				t.Errorf("Expected remaining = %d, got %d", expectedRemaining, remaining)
			}
		}
	})

	t.Run("Block requests exceeding limit", func(t *testing.T) {
		ip := "192.168.1.101"
		limit := 3
		window := time.Hour

		// Use up the limit
		for i := 0; i < 3; i++ {
			allowed, _, _, _ := limiter.AllowIP(ip, limit, window)
			if !allowed {
				t.Errorf("Request %d should be allowed", i+1)
			}
		}

		// 4th request should be blocked
		allowed, _, remaining, _ := limiter.AllowIP(ip, limit, window)
		if allowed {
			t.Error("Request should be blocked after exceeding limit")
		}
		if remaining != 0 {
			t.Errorf("Expected remaining = 0, got %d", remaining)
		}
	})

	t.Run("Different IPs have independent limits", func(t *testing.T) {
		ip1 := "192.168.1.102"
		ip2 := "192.168.1.103"
		limit := 2
		window := time.Hour

		// Use up ip1's limit
		for i := 0; i < 2; i++ {
			limiter.AllowIP(ip1, limit, window)
		}

		// ip1 should be blocked
		allowed, _, _, _ := limiter.AllowIP(ip1, limit, window)
		if allowed {
			t.Error("IP1 should be blocked")
		}

		// ip2 should still be allowed
		allowed, _, _, _ = limiter.AllowIP(ip2, limit, window)
		if !allowed {
			t.Error("IP2 should be allowed (independent limit)")
		}
	})

	t.Run("Reset time is calculated correctly", func(t *testing.T) {
		ip := "192.168.1.104"
		limit := 5
		window := time.Hour

		before := time.Now()
		_, _, _, resetAt := limiter.AllowIP(ip, limit, window)
		after := time.Now()

		// Reset should be approximately 1 hour from now
		expectedReset := before.Add(window)
		timeDiff := resetAt.Sub(expectedReset)

		// Allow up to 1 second tolerance for test execution time
		if timeDiff < 0 || timeDiff > time.Second {
			t.Errorf("Reset time not correct. Expected ~%v, got %v (diff: %v)",
				expectedReset, resetAt, timeDiff)
		}

		// Reset should be after the current time
		if !resetAt.After(after) {
			t.Error("Reset time should be in the future")
		}
	})

	t.Run("Window reset allows new requests", func(t *testing.T) {
		ip := "192.168.1.105"
		limit := 2
		window := 100 * time.Millisecond // Very short window for testing

		// Use up the limit
		for i := 0; i < 2; i++ {
			limiter.AllowIP(ip, limit, window)
		}

		// Should be blocked
		allowed, _, _, _ := limiter.AllowIP(ip, limit, window)
		if allowed {
			t.Error("Should be blocked before window reset")
		}

		// Wait for window to reset
		time.Sleep(150 * time.Millisecond)

		// Should be allowed again after reset
		allowed, _, remaining, _ := limiter.AllowIP(ip, limit, window)
		if !allowed {
			t.Error("Should be allowed after window reset")
		}
		if remaining != limit-1 {
			t.Errorf("Expected remaining = %d after reset, got %d", limit-1, remaining)
		}
	})

	t.Run("Custom window durations work correctly", func(t *testing.T) {
		// Use a unique IP to avoid interference from other tests
		newLimiter := NewLimiter() // Fresh limiter for this test
		ip := "10.0.0.1"

		// Test with 30 minute window
		limit := 10
		window := 30 * time.Minute

		before := time.Now()
		allowed, limitVal, _, resetAt := newLimiter.AllowIP(ip, limit, window)
		after := time.Now()

		if !allowed {
			t.Error("First request should be allowed")
		}
		if limitVal != limit {
			t.Errorf("Expected limit = %d, got %d", limit, limitVal)
		}

		// Reset should be 30 minutes from now (approximately)
		expectedReset := before.Add(window)
		timeDiff := resetAt.Sub(expectedReset)
		// Allow more tolerance for window execution time
		if timeDiff < 0 || timeDiff > 2*time.Second {
			t.Errorf("Reset time not aligned with custom window. Expected ~%v, got %v (diff: %v)",
				expectedReset, resetAt, timeDiff)
		}

		// Verify it's in the future
		if !resetAt.After(after) {
			t.Error("Reset time should be in the future")
		}
	})
}

func TestAllowIP_Concurrent(t *testing.T) {
	limiter := NewLimiter()
	ip := "192.168.1.200"
	limit := 100
	window := time.Hour

	// Simulate concurrent requests (should handle race conditions)
	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			limiter.AllowIP(ip, limit, window)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Next request should be blocked (all 100 used up)
	allowed, _, remaining, _ := limiter.AllowIP(ip, limit, window)
	if allowed {
		t.Error("Should be blocked after concurrent requests exhausted limit")
	}
	if remaining != 0 {
		t.Errorf("Expected remaining = 0, got %d", remaining)
	}
}
