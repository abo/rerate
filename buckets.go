package rerate

import "time"

// Buckets a set of bucket, each bucket compute and return the number of occurs in itself
type Buckets interface {
	Inc(key string, id int64) error
	Del(key string, ids ...int64) error
	Get(key string, ids ...int64) ([]int64, error)
}

// BucketsFactory a interface to create Buckets
type BucketsFactory func(size int64, ttl time.Duration) Buckets
