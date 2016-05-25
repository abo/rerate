package ratelimiter

import (
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
)

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

func TestInc(t *testing.T) {
	pool := newPool("localhost:6379", "")
	counter := NewCounter(pool, "ratelimiter:test", time.Minute, time.Second)
	ip1, ip2 := "114.255.86.248", "114.255.86.249"

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
