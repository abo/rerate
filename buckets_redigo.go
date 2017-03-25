package rerate

import (
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedigoBuckets is a Buckets using redigo as backend
type RedigoBuckets struct {
	pool *redis.Pool
	size int64
	ttl  time.Duration
}

// NewRedigoBuckets is a RedigoBuckets factory
func NewRedigoBuckets(redis *redis.Pool) BucketsFactory {
	return func(size int64, ttl time.Duration) Buckets {
		return &RedigoBuckets{
			pool: redis,
			size: size,
			ttl:  ttl,
		}
	}

}

// cleanup unused bucket(s)
func (bs *RedigoBuckets) cleanup(key string, from int64) {
	conn := bs.pool.Get()
	defer conn.Close()

	if l, err := redis.Int64(conn.Do("HLEN", key)); err != nil || l < bs.size*2 {
		return
	}

	if ids, err := redis.Strings(conn.Do("HKEYS", key)); err == nil {
		var delIds []int64
		for _, s := range ids {
			if v, err := strconv.ParseInt(s, 10, 64); err == nil && v <= from-bs.size {
				delIds = append(delIds, v)
			}
		}
		bs.Del(key, delIds...)
	}
}

// Inc increment bucket key:id 's occurs
func (bs *RedigoBuckets) Inc(key string, id int64) error {
	conn := bs.pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	conn.Send("HINCRBY", key, strconv.FormatInt(id, 10), 1)
	conn.Send("PEXPIRE", key, int64(bs.ttl/time.Millisecond))
	ret, err := redis.Ints(conn.Do("EXEC"))
	if err != nil {
		return err
	}

	if ret[0] == 1 { // new bucket created
		go bs.cleanup(key, id)
	}
	return nil
}

// Del delete bucket key:ids, or delete Buckets key when ids is empty.
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

// Get return bucket key:ids' occurs
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
