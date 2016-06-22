package rerate_test

import (
	"fmt"
	"time"

	"github.com/abo/rerate"
	"github.com/garyburd/redigo/redis"
)

func newRedisPool(server, password string) *redis.Pool {
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

func ExampleCounter() {
	redis := newRedisPool("localhost:6379", "")
	key := "pv-home"
	// pv count in 5s, try to release per 0.5s
	counter := rerate.NewCounter(redis, "rr:test:count", 5*time.Second, 500*time.Millisecond)

	ticker := time.NewTicker(1000 * time.Millisecond)
	go func() {
		for _ = range ticker.C {
			counter.Inc(key)
		}
	}()

	time.Sleep(4500 * time.Millisecond)
	ticker.Stop()
	total, _ := counter.Count(key)
	his, _ := counter.Histogram(key)
	fmt.Println("total:", total, ", histogram:", his)
	//Output: total: 4 , histogram: [0 1 0 1 0 1 0 1 0 0]
}

func ExampleLimiter() {
	redis := newRedisPool("localhost:6379", "")
	key := "pv-dashboard"
	// rate limit to 10/2s, release interval 0.2s
	limiter := rerate.NewLimiter(redis, "rr:test:limit", 2*time.Second, 200*time.Millisecond, 10)

	ticker := time.NewTicker(200 * time.Millisecond)
	go func() {
		for _ = range ticker.C {
			limiter.Inc(key)
		}
	}()

	time.Sleep(1850 * time.Millisecond)
	rem, _ := limiter.Remaining(key)
	exceed, _ := limiter.Exceeded(key)
	fmt.Println(rem, exceed) // after 9 ticks, count 9, remaining 1

	time.Sleep(200 * time.Millisecond)
	ticker.Stop()
	rem, _ = limiter.Remaining(key)
	exceed, _ = limiter.Exceeded(key)
	fmt.Println(rem, exceed) // after 10 ticks, count 10, remaining 0, exceeded

	time.Sleep(150 * time.Millisecond)
	rem, _ = limiter.Remaining(key)
	exceed, _ = limiter.Exceeded(key)
	fmt.Println(rem, exceed) // first inc should be released, remaining 1
	//Output: 1 false
	// 0 true
	// 1 false
}
