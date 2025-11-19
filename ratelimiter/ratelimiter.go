package ratelimiter

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements a token bucket algorithm for rate limiting.
type RateLimiter struct {
	visitors sync.Map
	rate     int64         // requests per window
	window   time.Duration // time window
}

var _ fmt.Stringer = &RateLimiter{}

type visitor struct {
	sync.Mutex

	tokens     int64
	lastUpdate int64
}

var _ fmt.Stringer = &visitor{}

// String implements fmt.Stringer.
func (v *visitor) String() string {
	v.Lock()
	defer v.Unlock()
	return fmt.Sprintf("%d:%d", v.tokens, v.lastUpdate)
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

// String implements fmt.Stringer.
func (rl *RateLimiter) String() string {
	visitors := make([]string, 0)
	rl.visitors.Range(func(key, value any) bool {
		visitors = append(visitors, fmt.Sprintf("%s: %s", key, value))
		return true
	})
	return strings.Join(visitors, ", ")
}

// Middleware returns an HTTP middleware that applies rate limiting.
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

		allow, tokens, lastUpdate := rl.allow(ip)
		resetTime := time.Unix(0, lastUpdate).Add(rl.window)

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.rate))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", tokens))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allow {
			retryAfter := int(math.Max(0, time.Until(resetTime).Seconds()))
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))

			http.Error(w, fmt.Sprintf("Rate limit exceeded; try again in %d seconds", retryAfter), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getVisitor retrieves or creates a visitor entry.
func (rl *RateLimiter) getVisitor(ip string) *visitor {
	if v, exists := rl.visitors.Load(ip); exists {
		return v.(*visitor)
	}
	v := &visitor{
		tokens:     rl.rate,
		lastUpdate: time.Now().UnixNano(),
	}

	actual, _ := rl.visitors.LoadOrStore(ip, v)
	return actual.(*visitor)
}

// allow checks if a request should be allowed.
func (rl *RateLimiter) allow(ip string) (bool, int64, int64) {
	now := time.Now().UnixNano()
	v := rl.getVisitor(ip)

	v.Lock()
	defer v.Unlock()

	for {
		elapsed := time.Duration(now - v.lastUpdate)

		if elapsed >= rl.window {
			v.lastUpdate = now
			v.tokens = rl.rate - 1
			return true, v.tokens, v.lastUpdate
		}

		if v.tokens > 0 {
			v.tokens = v.tokens - 1
			return true, v.tokens, v.lastUpdate
		}

		return false, v.tokens, v.lastUpdate
	}
}

// cleanupVisitors removes old visitor entries to prevent memory leaks.
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-rl.window * 2).UnixNano()

		rl.visitors.Range(func(key, value any) bool {
			v := value.(*visitor)
			v.Lock()
			if v.lastUpdate < cutoff {
				rl.visitors.Delete(key)
			}
			v.Unlock()
			return true
		})
	}
}
