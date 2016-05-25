package ratelimiter

import (
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	pool := newPool("localhost:6379", "")
	limiter := NewLimiter(pool, "ratelimiter:test", time.Minute, time.Second, 20)
	for i := 0; i < 19; i++ {
		if rem, err := limiter.Remaining("abc"); err != nil || rem != int64(20-i) {
			t.Fatal("remaining should be ", 20-i, ", but actual ", rem)
		}
		limiter.Inc("abc")
	}
	if exceed, err := limiter.Exceeded("abc"); err != nil || exceed {
		t.Fatal("should not exceeded", err)
	}

	limiter.Inc("abc")
	if exceed, err := limiter.Exceeded("abc"); err != nil || !exceed {
		t.Fatal("should exceeded", err)
	}
}
