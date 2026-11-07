package larapak

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientLimit struct {
	count     int
	resetTime time.Time
}

// RateLimiter tracks request frequency per client IP.
type RateLimiter struct {
	mu      sync.Mutex
	limits  map[string]*clientLimit
	cleanup time.Duration
}

var (
	defaultLimiter     *RateLimiter
	defaultLimiterOnce sync.Once
)

// DefaultRateLimiter returns the singleton in-memory RateLimiter instance.
func DefaultRateLimiter() *RateLimiter {
	defaultLimiterOnce.Do(func() {
		defaultLimiter = &RateLimiter{
			limits:  make(map[string]*clientLimit),
			cleanup: 1 * time.Minute,
		}
		go defaultLimiter.startJanitor()
	})
	return defaultLimiter
}

func (rl *RateLimiter) startJanitor() {
	ticker := time.NewTicker(rl.cleanup)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, cl := range rl.limits {
			if now.After(cl.resetTime) {
				delete(rl.limits, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if the given client IP is allowed to proceed under the quota.
func (rl *RateLimiter) Allow(ip string, maxRequests int, window time.Duration) (bool, int, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cl, exists := rl.limits[ip]
	if !exists || now.After(cl.resetTime) {
		rl.limits[ip] = &clientLimit{
			count:     1,
			resetTime: now.Add(window),
		}
		return true, maxRequests - 1, window
	}

	if cl.count >= maxRequests {
		retryAfter := time.Until(cl.resetTime)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, 0, retryAfter
	}

	cl.count++
	remaining := maxRequests - cl.count
	return true, remaining, time.Until(cl.resetTime)
}

// Throttle creates a rate limiting middleware enforcing maxRequests per time window.
func Throttle(maxRequests int, window time.Duration) MiddlewareFunc {
	limiter := DefaultRateLimiter()

	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			ip := c.Request.RemoteAddr
			if forwarded := c.Request.Header.Get("X-Forwarded-For"); forwarded != "" {
				parts := strings.Split(forwarded, ",")
				ip = strings.TrimSpace(parts[0])
			} else if realIP := c.Request.Header.Get("X-Real-IP"); realIP != "" {
				ip = realIP
			}

			allowed, remaining, retryAfter := limiter.Allow(ip, maxRequests, window)

			c.Writer.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
			c.Writer.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			if !allowed {
				retrySec := int(retryAfter.Seconds())
				if retrySec < 1 {
					retrySec = 1
				}
				c.Writer.Header().Set("Retry-After", fmt.Sprintf("%d", retrySec))

				accept := c.Request.Header.Get("Accept")
				if strings.Contains(accept, "application/json") || strings.HasPrefix(c.Request.URL.Path, "/api") {
					return c.JSON(http.StatusTooManyRequests, map[string]any{
						"status":      429,
						"error":       "Kakehan panjalukan (Too Many Requests). Tulung enteni sedhela.",
						"retry_after": retrySec,
					})
				}

				c.Writer.WriteHeader(http.StatusTooManyRequests)
				_, _ = c.Writer.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>429 - Kakehan Panjalukan</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0B0C10; color: #F8FAFC; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
        .card { background: #111218; border: 1px solid #1E202B; padding: 32px; border-radius: 12px; max-width: 480px; text-align: center; }
        h1 { font-size: 24px; color: #EF4444; margin-bottom: 12px; }
        p { color: #94A3B8; font-size: 14px; line-height: 1.5; margin-bottom: 20px; }
        .badge { display: inline-block; background: #1E202B; color: #38BDF8; padding: 4px 12px; border-radius: 6px; font-family: monospace; font-size: 13px; }
    </style>
</head>
<body>
    <div class="card">
        <h1>429 - Kakehan Panjalukan</h1>
        <p>Aktivitas panjalukan sampeyan ngluwihi wates sing ditemtokna. Kanggo njaga stabilitas server, tulung enteni sedhela sakdurunge nyoba maning.</p>
        <span class="badge">Retry-After: ` + fmt.Sprintf("%d", retrySec) + ` detik</span>
    </div>
</body>
</html>`))
				return nil
			}

			return next(c)
		}
	}
}
