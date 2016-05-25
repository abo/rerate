package ratelimiter

import (
	"bytes"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Counter count total occurs during a time window,
// it will store occurs during every time slice: (now ~ now - step), (now - step ~ now - 2*step)...
type Counter struct {
	pool   Pool
	prefix string
	window time.Duration
	step   time.Duration

	buckets int
}

// NewCounter create a new Counter
func NewCounter(pool Pool, prefix string, window, step time.Duration) *Counter {
	return &Counter{
		pool:    pool,
		prefix:  prefix,
		window:  window,
		step:    step,
		buckets: int(window/step) + 1,
	}
}

func (c *Counter) bucket() int {
	now := time.Now().UnixNano()
	return int(now/int64(c.step)) % c.buckets
}

func (c *Counter) key(id string) string {
	buf := bytes.NewBufferString(c.prefix)
	buf.WriteString(":")
	buf.WriteString(id)
	return buf.String()
}

// Inc increment id's occurs, id is the event, may be user's ip, userid ...
func (c *Counter) Inc(id string) error {
	bucket := c.bucket()

	conn := c.pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	conn.Send("HINCRBY", c.key(id), strconv.Itoa(bucket), 1)
	conn.Send("HDEL", c.key(id), strconv.Itoa(bucket+1))
	conn.Send("PEXPIRE", c.key(id), c.window/time.Millisecond)
	_, err := conn.Do("EXEC")

	return err
}

// Count return total occurs in duration(buckets * precision)
func (c *Counter) Count(id string) (int64, error) {
	bucket := c.bucket()
	args := make([]interface{}, c.buckets)
	args[0] = c.key(id)
	for i := 0; i < c.buckets-1; i++ {
		args[i+1] = strconv.Itoa((c.buckets + bucket - i) % c.buckets)
	}

	conn := c.pool.Get()
	defer conn.Close()

	vals, err := redis.Strings(conn.Do("HMGET", args...))
	if err != nil {
		return 0, err
	}

	total := int64(0)
	for _, val := range vals {
		if v, e := strconv.ParseInt(val, 10, 64); e == nil {
			total += v
		}
	}

	return total, nil
}

// Reset cleanup occurs, set it to zero
func (c *Counter) Reset(id string) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", c.key(id))
	return err
}
