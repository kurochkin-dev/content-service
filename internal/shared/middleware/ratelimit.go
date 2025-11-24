package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	RateLimitTokens = 100
	RateLimitRefill = time.Second / 10
	CleanupInterval = 10 * time.Minute
	LimiterTTL      = 30 * time.Minute
)

type rateLimiter struct {
	tokens         int
	maxTokens      int
	refillRate     time.Duration
	lastRefillTime time.Time
	lastAccessTime time.Time
	mu             sync.Mutex
}

func newRateLimiter(maxTokens int, refillRate time.Duration) *rateLimiter {
	now := time.Now()
	return &rateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: now,
		lastAccessTime: now,
	}
}

func (rl *rateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.lastAccessTime = now
	elapsed := now.Sub(rl.lastRefillTime)

	tokensToAdd := int(elapsed / rl.refillRate)
	if tokensToAdd > 0 {
		rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
		rl.lastRefillTime = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

type rateLimiterStore struct {
	limiters map[string]*rateLimiter
	mu       sync.RWMutex
}

func newRateLimiterStore() *rateLimiterStore {
	store := &rateLimiterStore{
		limiters: make(map[string]*rateLimiter),
	}

	go store.cleanup()

	return store
}

func (s *rateLimiterStore) getLimiter(ip string) *rateLimiter {
	s.mu.RLock()
	limiter, exists := s.limiters[ip]
	s.mu.RUnlock()

	if exists {
		return limiter
	}

	s.mu.Lock()
	limiter, exists = s.limiters[ip]
	if !exists {
		limiter = newRateLimiter(RateLimitTokens, RateLimitRefill)
		s.limiters[ip] = limiter
	}
	s.mu.Unlock()

	return limiter
}

func (s *rateLimiterStore) cleanup() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		removed := 0

		for ip, limiter := range s.limiters {
			limiter.mu.Lock()
			if now.Sub(limiter.lastAccessTime) > LimiterTTL {
				delete(s.limiters, ip)
				removed++
			}
			limiter.mu.Unlock()
		}

		s.mu.Unlock()

		if removed > 0 {
			log.Debug().Int("removed", removed).Int("remaining", len(s.limiters)).Msg("Cleaned up inactive rate limiters")
		}
	}
}

var store = newRateLimiterStore()

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := store.getLimiter(ip)

		if !limiter.allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
