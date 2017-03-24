package rerate

import "time"

type Buckets interface {
	Inc(key string, id int64) error
	Del(key string, ids ...int64) error
	Get(key string, ids ...int64) ([]int64, error)
}

type BucketsFactory func(size int64, ttl time.Duration) Buckets
