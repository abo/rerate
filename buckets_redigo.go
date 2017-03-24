package rerate

import (
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

type RedigoBuckets struct {
	pool *redis.Pool
	size int64
	ttl  time.Duration
}

func NewRedigoBuckets(redis *redis.Pool) BucketsFactory {
	return func(size int64, ttl time.Duration) Buckets {
		return &RedigoBuckets{
			pool: redis,
			size: size,
			ttl:  ttl,
		}
	}

	// p := &redis.Pool{
	// 	MaxIdle:     3,
	// 	IdleTimeout: 240 * time.Second,
	// 	Dial: func() (redis.Conn, error) {
	// 		c, err := redis.Dial("tcp", server)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		if len(password) == 0 {
	// 			return c, err
	// 		}

	// 		if _, err := c.Do("AUTH", password); err != nil {
	// 			c.Close()
	// 			return nil, err
	// 		}
	// 		return c, err
	// 	},
	// 	TestOnBorrow: func(c redis.Conn, t time.Time) error {
	// 		_, err := c.Do("PING")
	// 		return err
	// 	},
	// }

}

func (bs *RedigoBuckets) gc(key string, id int64) {
	conn := bs.pool.Get()
	defer conn.Close()

	l, _ := redis.Int64(conn.Do("HLEN", key))
	if l >= bs.size*2 {
		// := redis.Strings(conn.Do("HKEYS", key))

		fmt.Println("NEED GC:", l, ">=", bs.size)
	}
}

func (bs *RedigoBuckets) Inc(key string, id int64) error {
	// inc id,  clear (id+1, id+2, id+3 ... , id+size/2 ), expire all after ttl
	conn := bs.pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	conn.Send("HINCRBY", key, strconv.FormatInt(id, 10), 1)
	conn.Send("PEXPIRE", key, int64(bs.ttl/time.Millisecond))
	_, err := conn.Do("EXEC")
	go bs.gc(key, id)
	return err
}

func (bs *RedigoBuckets) Del(key string, ids ...int64) error {
	conn := bs.pool.Get()
	defer conn.Close()

	if len(ids) == 0 {
		_, err := conn.Do("DEL", key)
		return err
	}

	args := make([]interface{}, len(ids)+1)
	args[0] = key
	for i, v := range ids {
		args[i+1] = v
	}
	_, err := conn.Do("HDEL", args...)
	return err
}

func (bs *RedigoBuckets) Get(key string, ids ...int64) ([]int64, error) {
	args := make([]interface{}, len(ids)+1)
	args[0] = key
	for i, v := range ids {
		args[i+1] = v
	}

	conn := bs.pool.Get()
	defer conn.Close()

	vals, err := redis.Strings(conn.Do("HMGET", args...))
	if err != nil {
		return []int64{}, err
	}

	ret := make([]int64, len(ids))
	for i, val := range vals {
		if v, e := strconv.ParseInt(val, 10, 64); e == nil {
			ret[i] = v
		} else {
			ret[i] = 0
		}
	}
	return ret, nil
}
