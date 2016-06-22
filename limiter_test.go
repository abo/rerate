package rerate_test

import (
	"testing"
	"time"

	. "github.com/abo/rerate"
)

func TestLimiter(t *testing.T) {
	limiter := NewLimiter(pool, "rerate:test:limiter:limiter", time.Minute, time.Second, 20)
	k := randkey()
	limiter.Reset(k)

	assertExceeded(t, limiter, k, false)
	for i := 0; i < 19; i++ {
		assertRem(t, limiter, k, int64(20-i))
		limiter.Inc(k)
		assertExceeded(t, limiter, k, false)
	}

	limiter.Inc(k)
	assertExceeded(t, limiter, k, true)
}

func TestExpire(t *testing.T) {
	limiter := NewLimiter(pool, "rerate:test:limiter:expire", 3*time.Second, time.Second, 20)
	k := randkey()
	limiter.Reset(k)

	limiter.Inc(k)
	assertRem(t, limiter, k, 19)

	wait(time.Second)
	limiter.Inc(k)
	assertRem(t, limiter, k, 18)

	wait(2 * time.Second)
	assertRem(t, limiter, k, 19)

	wait(time.Second)
	assertRem(t, limiter, k, 20)
}

//TODO 测试period不是interval的整数倍

func TestNonOccurs(t *testing.T) {
	l := NewLimiter(pool, "rerate:test:limiter:nonoccurs", 3*time.Second, 500*time.Millisecond, 20)
	k := randkey()
	l.Reset(k)
	assertRem(t, l, k, 20)

	for i := 0; i < 6; i++ {
		l.Inc(k)
		wait(500 * time.Millisecond)
	}
	assertRem(t, l, k, 15)

	for i := 0; i < 5; i++ {
		wait(500 * time.Millisecond)
		assertRem(t, l, k, int64(15+1+i))
	}
}

func assertRem(t *testing.T, l *Limiter, k string, expect int64) {
	if c, err := l.Remaining(k); err != nil || c != expect {
		t.Fatal("expect ", expect, " actual ", c, ", err:", err)
	}
}

func assertExceeded(t *testing.T, l *Limiter, k string, expect bool) {
	if exceed, err := l.Exceeded(k); err != nil || exceed != expect {
		t.Fatal("expect exceeded:", expect, ",err ", err)
	}
}
