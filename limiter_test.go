package rerate

import (
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	limiter := NewLimiter(pool, "rerate:test:limiter:limiter", time.Minute, time.Second, 20)
	k := randkey()
	limiter.Reset(k)

	if exceed, err := limiter.Exceeded(k); err != nil || exceed {
		t.Fatal("should not exceeded", err)
	}
	for i := 0; i < 19; i++ {
		if rem, err := limiter.Remaining(k); err != nil || rem != int64(20-i) {
			t.Fatal("remaining should be ", 20-i, ", but actual ", rem)
		}

		limiter.Inc(k)

		if exceed, err := limiter.Exceeded(k); err != nil || exceed {
			t.Fatal("should not exceeded", err)
		}
	}

	limiter.Inc(k)
	if exceed, err := limiter.Exceeded(k); err != nil || !exceed {
		t.Fatal("should exceeded", err)
	}
}

func TestExpire(t *testing.T) {
	limiter := NewLimiter(pool, "rerate:test:limiter:expire", 3*time.Second, time.Second, 20)
	k := randkey()
	limiter.Reset(k)

	limiter.Inc(k)
	if c, _ := limiter.Remaining(k); c != 19 {
		t.Fatal("expect 19, remaining ", c)
	}

	wait(time.Second)
	limiter.Inc(k)
	if c, _ := limiter.Remaining(k); c != 18 {
		t.Fatal("expect 18, remaining ", c)
	}

	wait(2 * time.Second)
	if c, _ := limiter.Remaining(k); c != 19 {
		t.Fatal("expect 1 released", c)
	}

	wait(time.Second)
	if c, _ := limiter.Remaining(k); c != 20 {
		t.Fatal("expect all released", c)
	}
}
