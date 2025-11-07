package ratelimiter

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// RateLimiter implements a token bucket algorithm for rate limiting
type RateLimiter struct {
	visitors sync.Map
	rate     int64         // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens     atomic.Int64
	lastUpdate atomic.Int64
}

// NewRateLimiter creates a new rate limiter
// rate: max requests per window
// window: time window for rate limiting
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rate:   int64(rate),
		window: window,
	}

	// Cleanup old visitors every 5 minutes
	go rl.cleanupVisitors()

	return rl
}

// getVisitor retrieves or creates a visitor entry
func (rl *RateLimiter) getVisitor(ip string) *visitor {
	if v, exists := rl.visitors.Load(ip); exists {
		return v.(*visitor)
	}
	v := &visitor{}
	v.tokens.Store(rl.rate)
	v.lastUpdate.Store(time.Now().UnixNano())

	actual, _ := rl.visitors.LoadOrStore(ip, v)
	return actual.(*visitor)
}

// allow checks if a request should be allowed
func (rl *RateLimiter) allow(ip string) bool {
	v := rl.getVisitor(ip)

	now := time.Now().UnixNano()

	for {
		lastUpdate := v.lastUpdate.Load()
		elapsed := time.Duration(now - lastUpdate)

		if elapsed >= rl.window {
			if v.lastUpdate.CompareAndSwap(lastUpdate, now) {
				v.tokens.Store(rl.rate - 1)
				return true
			}
			// CAS failed, another goroutine updated it, try again
			continue
		}

		tokens := v.tokens.Load()
		if tokens > 0 {
			if v.tokens.CompareAndSwap(tokens, tokens-1) {
				return true
			}
			// CAS failed, another goroutine took a token, try again
			continue
		}

		return false
	}
}

// cleanupVisitors removes old visitor entries to prevent memory leaks
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-rl.window * 2).UnixNano()

		rl.visitors.Range(func(key, value any) bool {
			v := value.(*visitor)
			if v.lastUpdate.Load() < cutoff {
				rl.visitors.Delete(key)
			}
			return true
		})
	}
}

// Middleware returns an HTTP middleware that applies rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract IP from request
		var ip string
		xff := r.Header["X-Forwarded-For"]
		if len(xff) > 0 {
			ip = xff[0]
		} else {
			if remoteIP, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
				ip = r.RemoteAddr
			} else {
				ip = remoteIP
			}
		}

		v := rl.getVisitor(ip)
		tokens := v.tokens.Load()
		lastUpdate := v.lastUpdate.Load()
		resetTime := time.Unix(0, lastUpdate).Add(rl.window)

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.rate))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(tokens-1, 0)))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !rl.allow(ip) {
			retryAfter := int(time.Until(resetTime).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))

			http.Error(w, fmt.Sprintf("Rate limit exceeded; try again in %d seconds", retryAfter), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
