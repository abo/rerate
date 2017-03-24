package rerate

import "time"

// Limiter a redis-based ratelimiter
type Limiter struct {
	Counter
	max int64
}

// NewLimiter create a new redis-based ratelimiter
// the Limiter limits the rate to max times per period
func NewLimiter(newBuckets BucketsFactory, pfx string, period, interval time.Duration, max int64) *Limiter {
	return &Limiter{
		Counter: *NewCounter(newBuckets, pfx, period, interval),
		max:     max,
	}
}

func (l *Limiter) remainingAt(id string, t time.Time) (int64, error) {
	occurs, err := l.countAt(id, t)
	if err != nil {
		return 0, err
	}
	return l.max - occurs, nil
}

// Remaining return the number of requests left for the time window
func (l *Limiter) Remaining(id string) (int64, error) {
	return l.remainingAt(id, time.Now())
}

func (l *Limiter) exceededAt(id string, t time.Time) (bool, error) {
	rem, err := l.Remaining(id)
	if err != nil {
		return false, err
	}
	return rem <= 0, nil
}

// Exceeded is exceeded the rate limit or not
func (l *Limiter) Exceeded(id string) (bool, error) {
	return l.exceededAt(id, time.Now())
}
