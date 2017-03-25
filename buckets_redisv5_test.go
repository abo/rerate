package rerate_test

import (
	"testing"
	"time"

	"github.com/abo/rerate"

	redis "gopkg.in/redis.v5"
)

func TestCleanup(t *testing.T) {
	buckets := rerate.NewRedisV5Buckets(redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}))(5, 5*time.Second)
	key := "buckets:redisv5:cleanup"

	for index := 0; index < 10; index++ {
		buckets.Inc(key, int64(index))
	}
	//TODO buckets's len == 5
}
