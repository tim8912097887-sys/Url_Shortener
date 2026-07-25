package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	capacity       int
	tokens         float64
	tokenFillRate  int // tokens per second 
	lastFilledTime time.Time
	mu          sync.Mutex
}

func NewRateLimiter(capacity int, tokenFillRate int) *RateLimiter {
	return &RateLimiter{capacity: capacity, tokens: float64(capacity), tokenFillRate: tokenFillRate, lastFilledTime: time.Now(), mu: sync.Mutex{}}
}

func (r *RateLimiter) refill() {
	now := time.Now()
	timeInterval := now.Sub(r.lastFilledTime).Seconds()
	r.tokens += timeInterval * float64(r.tokenFillRate)
	if r.tokens > float64(r.capacity) {
		r.tokens = float64(r.capacity)
	}
	r.lastFilledTime = now
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
    r.refill()
	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

func (r *RateLimiter) AllowN(n int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refill()
	if r.tokens >= float64(n) {
		r.tokens -= float64(n)
		return true
	}
	return false
}