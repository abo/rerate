package ratelimiter

import "time"

// Limiter a redis-based ratelimiter
type Limiter struct {
	Counter
	max int64
}

// NewLimiter create a new redis-based ratelimiter
func NewLimiter(pool Pool, prefix string, window, step time.Duration, max int64) *Limiter {
	return &Limiter{
		Counter: *NewCounter(pool, prefix, window, step),
		max:     max,
	}
}

// Remaining return the number of requests left for the time window
func (l *Limiter) Remaining(id string) (int64, error) {
	occurs, err := l.Count(id)
	if err != nil {
		return 0, err
	}
	return l.max - occurs, nil
}

// Exceeded exceeded the rate limit or not
func (l *Limiter) Exceeded(id string) (bool, error) {
	rem, err := l.Remaining(id)
	if err != nil {
		return false, err
	}
	return rem <= 0, nil
}
