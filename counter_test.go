package rerate

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
)

var pool Pool

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if len(password) == 0 {
				return c, err
			}

			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func wait(duration time.Duration) {
	t1 := time.NewTimer(duration)
	<-t1.C
}

func randkey() string {
	return strconv.Itoa(rand.Int())
}

func init() {
	pool = newPool("localhost:6379", "")
}

func TestBuckets(t *testing.T) {
	testcases := map[int][]int{
		1:  {1, 0, 19, 18, 17, 16, 15, 14, 13, 12},
		0:  {0, 19, 18, 17, 16, 15, 14, 13, 12, 11},
		10: {10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	}

	counter := NewCounter(pool, "rerate:test:counter:buckets", 10*time.Second, time.Second)
	for input, expect := range testcases {
		buckets := counter.buckets(input)
		if !reflect.DeepEqual(buckets, expect) {
			t.Fatal("expect ", expect, ",but ", buckets)
		}
	}
}

func TestHash(t *testing.T) {
	// hash(a+s) = hash(a)+1
	counter := NewCounter(pool, "rerate:test:counter:hash", time.Minute, time.Second)
	l := int(time.Minute / time.Second)

	testcases := []int64{time.Now().UnixNano(),
		time.Now().UnixNano() - int64(rand.Intn(100))*int64(time.Second),
		time.Now().UnixNano() - int64(rand.Intn(100))*int64(time.Second),
		time.Now().UnixNano() - int64(rand.Intn(100))*int64(time.Second),
		time.Now().UnixNano() - int64(rand.Intn(100))*int64(time.Second)}

	for _, input := range testcases {
		next := input + int64(time.Second)
		b := counter.hash(input)
		nb := counter.hash(next)
		if b < 0 || b > 2*l {
			t.Fatal("out of range ", input)
		}
		if nb < 0 || nb > 2*l {
			t.Fatal("out of range ", input)
		}
		if b+1 != nb && b-(2*l) != nb {
			t.Fatal("input ", input)
		}
	}

}

func TestHistogram(t *testing.T) {
	counter := NewCounter(pool, "rerate:test:counter:count", 4000*time.Millisecond, 400*time.Millisecond)
	id := randkey()
	counter.Reset(id)
	assertHist(t, counter, id, []int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	current := counter.hash(time.Now().UnixNano())
	for i := 0; i <= current; i++ {
		for j := 0; j < i; j++ {
			counter.inc(id, i)
		}
	}
	hist, _ := counter.Histogram(id)

	for k := 0; k <= 10; k++ {
		wait(counter.interval)
		for i := len(hist) - 1; i > 0; i-- {
			hist[i] = hist[i-1]
		}
		hist[0] = 0

		assertHist(t, counter, id, hist)
	}
}

func TestCounter(t *testing.T) {
	counter := NewCounter(pool, "rerate:test:counter:counter", time.Minute, time.Second)
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

func assertHist(t *testing.T, c *Counter, k string, expect []int64) {
	if b, err := c.Histogram(k); err != nil || !reflect.DeepEqual(b, expect) {
		t.Fatal("expect ", expect, " without err, actual", b, err)
	}
}
