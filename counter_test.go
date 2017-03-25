package rerate_test

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	redis "gopkg.in/redis.v5"

	. "github.com/abo/rerate"
)

func randkey() string {
	return strconv.Itoa(rand.Int())
}

func TestHistogram(t *testing.T) {
	redisBuckets := NewRedisV5Buckets(redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}))
	counter := NewCounter(redisBuckets, "rerate:test:counter:count", 4000*time.Millisecond, 400*time.Millisecond)
	id := randkey()
	counter.Reset(id)

	zero := []int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if b, err := counter.Histogram(id); err != nil || !reflect.DeepEqual(b, zero) {
		t.Fatal("expect all zero without err, actual", b, err)
	}

	at := time.Now()
	for i := 0; i <= 10; i++ {
		for j := 0; j < i; j++ {
			IncAtExp(counter, id, at.Add(time.Duration(i)*400*time.Millisecond))
		}
	} //[]int64{0,1,2,3,4,5,6,7,8,9,10,0,0,0,0,0,0,0,0,0}

	for i := 0; i < 20; i++ {
		for j := len(zero) - 1; j > 0; j-- {
			zero[j] = zero[j-1]
		}
		if i <= 10 {
			zero[0] = int64(i)
		} else {
			zero[0] = 0
		}

		assertHist(t, counter, id, at.Add(time.Duration(i)*400*time.Millisecond), zero)
	}
}

func TestCounter(t *testing.T) {
	redisBuckets := NewRedisV5Buckets(redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}))
	counter := NewCounter(redisBuckets, "rerate:test:counter:counter", time.Minute, time.Second)
	ip1, ip2 := randkey(), randkey()

	if err := counter.Reset(ip1); err != nil {
		t.Fatal("can not reset counter", err)
	}
	if err := counter.Reset(ip2); err != nil {
		t.Fatal("can not reset counter", err)
	}

	assertCount(t, counter, ip1, 0)

	for i := 0; i < 10; i++ {
		counter.Inc(ip1)
		assertCount(t, counter, ip1, int64(i+1))
		assertCount(t, counter, ip2, 0)
	}
}

func assertCount(t *testing.T, c *Counter, k string, expect int64) {
	if count, err := c.Count(k); err != nil || count != expect {
		t.Fatal("should be ", expect, " without error, actual ", count, err)
	}
}

func assertHist(t *testing.T, c *Counter, k string, from time.Time, expect []int64) {
	if b, err := HistogramAtExp(c, k, from); err != nil || !reflect.DeepEqual(b, expect) {
		t.Fatal("expect ", expect, " without err, actual", b, err)
	}
}
