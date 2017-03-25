package rerate

import (
	"strconv"
	"time"

	redis "gopkg.in/redis.v5"
)

// RedisV5Buckets is a Buckets using redis.v5 as backend
type RedisV5Buckets struct {
	redis *redis.Client
	size  int64
	ttl   time.Duration
}

// NewRedisV5Buckets is a RedisV5Buckets factory
func NewRedisV5Buckets(redis *redis.Client) BucketsFactory {
	return func(size int64, ttl time.Duration) Buckets {
		return &RedisV5Buckets{
			redis: redis,
			size:  size,
			ttl:   ttl,
		}
	}
}

func (bs *RedisV5Buckets) cleanup(key string, from int64) {
	if l, err := bs.redis.HLen(key).Result(); err != nil || l < bs.size*2 {
		return
	}

	if ids, err := bs.redis.HKeys(key).Result(); err == nil {
		var delIds []int64
		for _, s := range ids {
			if v, e := strconv.ParseInt(s, 10, 64); e == nil && v <= from-bs.size {
				delIds = append(delIds, v)

			}
		}
		bs.Del(key, delIds...)
	}
}

// Inc increment bucket key:id 's occurs
func (bs *RedisV5Buckets) Inc(key string, id int64) error {
	pipe := bs.redis.TxPipeline()
	defer pipe.Close()

	count := pipe.HIncrBy(key, strconv.FormatInt(id, 10), 1)
	pipe.PExpire(key, bs.ttl)
	_, err := pipe.Exec()
	if err != nil {
		return err
	}

	if count.Val() == 1 { // new bucket created
		go bs.cleanup(key, id)
	}
	return nil
}

// Del delete bucket key:ids, or delete Buckets key when ids is empty.
func (bs *RedisV5Buckets) Del(key string, ids ...int64) error {
	if len(ids) == 0 {
		_, err := bs.redis.Del(key).Result()
		return err
	}

	args := make([]string, len(ids))
	for i, v := range ids {
		args[i] = strconv.FormatInt(v, 10)
	}
	_, err := bs.redis.HDel(key, args...).Result()
	return err
}

// Get return bucket key:ids' occurs
func (bs *RedisV5Buckets) Get(key string, ids ...int64) ([]int64, error) {
	args := make([]string, len(ids))
	for i, v := range ids {
		args[i] = strconv.FormatInt(v, 10)
	}

	results, err := bs.redis.HMGet(key, args...).Result()
	if err != nil {
		return []int64{}, err
	}

	vals := make([]int64, len(ids))
	for i, result := range results {
		if result == nil {
			vals[i] = 0
		} else if v, e := strconv.ParseInt(result.(string), 10, 64); e != nil {
			vals[i] = 0
		} else {
			vals[i] = v
		}
	}

	return vals, nil
}
