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

	for i := 0; i < 20; i++ {
		time.Sleep(200 * time.Millisecond)
		if exceed, _ := limiter.Exceeded(key); exceed {
			fmt.Println("exceeded")
			ticker.Stop()
		} else {
			rem, _ := limiter.Remaining(key)
			fmt.Println("remaining", rem)
		}
	}
	//Output:
	// remaining 9
	// remaining 8
	// remaining 7
	// remaining 6
	// remaining 5
	// remaining 4
	// remaining 3
	// remaining 2
	// remaining 1
	// exceeded
	// remaining 1
	// remaining 2
	// remaining 3
	// remaining 4
	// remaining 5
	// remaining 6
	// remaining 7
	// remaining 8
	// remaining 9
	// remaining 10
}
