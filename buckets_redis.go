package rerate

import (
	"strconv"
	"time"

	redis "gopkg.in/redis.v5"
)

type RedisBuckets struct {
	redis *redis.Client
	size  int64
	ttl   time.Duration
}

func NewRedisBuckets(redis *redis.Client) BucketsFactory {
	return func(size int64, ttl time.Duration) Buckets {
		return &RedisBuckets{
			redis: redis,
			size:  size,
			ttl:   ttl,
		}
	}
}

func (bs *RedisBuckets) Inc(key string, id int64) error {
	pipe := bs.redis.TxPipeline()
	defer pipe.Close()

	pipe.HIncrBy(key, strconv.FormatInt(id, 10), 1)
	pipe.PExpire(key, bs.ttl)
	_, err := pipe.Exec()
	return err
}

func (bs *RedisBuckets) Del(key string, ids ...int64) error {
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

func (bs *RedisBuckets) Get(key string, ids ...int64) ([]int64, error) {
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
