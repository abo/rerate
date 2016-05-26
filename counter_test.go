package ratelimiter

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
		1:  []int{1, 0, 10, 9, 8, 7, 6, 5, 4, 3},
		0:  []int{0, 10, 9, 8, 7, 6, 5, 4, 3, 2},
		10: []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	}

	counter := NewCounter(pool, "ratelimiter:test:buckets", 10*time.Second, time.Second)
	for input, expect := range testcases {
		buckets := counter.buckets(input)
		if !reflect.DeepEqual(buckets, expect) {
			t.Fatal("expect ", expect, ",but ", buckets)
		}
	}
}

func TestHash(t *testing.T) {
	// hash(a+s) = hash(a)+1
	counter := NewCounter(pool, "ratelimiter:test:hash", time.Minute, time.Second)
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
		if b < 0 || b > l {
			t.Fatal("out of range ", input)
		}
		if nb < 0 || nb > l {
			t.Fatal("out of range ", input)
		}
		if b+1 != nb && b-l != nb {
			t.Fatal("input ", input)
		}
	}

}

func TestCount(t *testing.T) {
	counter := NewCounter(pool, "ratelimiter:test:counter", 10*time.Second, time.Second)
	id := randkey()
	counter.Reset(id)
	// inc(id, 1) + inc(id, 2) = count(id)
	counter.inc(id, 0)
	counter.inc(id, 1)
	c, e := counter.count(id, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	if e != nil {
		t.Fatal(e)
	}
	if c != 2 {
		t.Fatal("expect 2, but ", c)
	}

	counter.inc(id, 1)
	c2, e2 := counter.count(id, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	if e2 != nil {
		t.Fatal(e2)
	}
	if c2 != 3 {
		t.Fatal("expect 3, but ", c2)
	}

	counter.inc(id, 0)
	c3, e3 := counter.count(id, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	if e3 != nil {
		t.Fatal(e3)
	}
	if c3 != 2 {
		t.Fatal("expect 2, but ", c3)
	}
}

func TestCounter(t *testing.T) {
	counter := NewCounter(pool, "ratelimiter:test:counter", time.Minute, time.Second)
	ip1, ip2 := randkey(), randkey()

	if err := counter.Reset(ip1); err != nil {
		t.Fatal("can not reset counter", err)
	}
	if err := counter.Reset(ip2); err != nil {
		t.Fatal("can not reset counter", err)
	}

	if c, err := counter.Count(ip1); err != nil || c != 0 {
		t.Fatal("should be 0 without error, ", c, err)
	}

	// if err := counter.Inc(ip1); err != nil {
	// 	t.Fatal("can not inc", ip1, err)
	// }
	// if c, err := counter.Count(ip1); err != nil || c != 1 {
	// 	t.Fatal("should be 1 without error, ", c, err)
	// }
	for i := 0; i < 10; i++ {
		counter.Inc(ip1)
	}

	if c, err := counter.Count(ip1); err != nil || c != 10 {
		t.Fatal("should be 11 without error, ", c, err)
	}

	if c, err := counter.Count(ip2); err != nil || c != 0 {
		t.Fatal("should be 0 without error, ", c, err)
	}
}
