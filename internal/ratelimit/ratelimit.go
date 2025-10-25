package ratelimit

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Limiter represents a rate limiter for a specific client
type Limiter struct {
	limit    int
	interval time.Duration
	tokens   int
	last     time.Time
	mutex    sync.Mutex
}

// NewLimiter creates a new rate limiter
func NewLimiter(limit int, interval time.Duration) *Limiter {
	return &Limiter{
		limit:    limit,
		interval: interval,
		tokens:   limit,
		last:     time.Now(),
	}
}

// Allow checks if a request is allowed based on rate limiting rules
func (l *Limiter) Allow() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	now := time.Now()
	// Add tokens based on time passed
	tokensToAdd := int(now.Sub(l.last).Seconds() / l.interval.Seconds() * float64(l.limit))
	if tokensToAdd > 0 {
		l.tokens += tokensToAdd
		if l.tokens > l.limit {
			l.tokens = l.limit
		}
		l.last = now
	}

	// Check if we have tokens available
	if l.tokens > 0 {
		l.tokens--
		return true
	}

	return false
}

// RateLimiter represents a collection of limiters for different clients
type RateLimiter struct {
	limiters map[string]*Limiter
	mutex    sync.Mutex
	limit    int
	interval time.Duration
}

// NewRateLimiter creates a new rate limiter collection
func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*Limiter),
		limit:    limit,
		interval: interval,
	}
}

// GetLimiter gets or creates a limiter for a specific key (IP address or API key)
func (rl *RateLimiter) GetLimiter(key string) *Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = NewLimiter(rl.limit, rl.interval)
		rl.limiters[key] = limiter
	}

	return limiter
}

// Middleware creates an HTTP middleware for rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use IP address as the key for rate limiting
		ip := getRealIP(r)
		limiter := rl.GetLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getRealIP extracts the real IP address from the request
func getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if there are multiple
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}