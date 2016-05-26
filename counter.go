package ratelimiter

import (
	"bytes"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Counter count total occurs during a time window w,
// it will store occurs during every time slice s: (now ~ now - s), (now - s ~ now - 2*s)...
type Counter struct {
	pool Pool
	pfx  string
	w    time.Duration
	s    time.Duration

	bkts int
}

// NewCounter create a new Counter
func NewCounter(pool Pool, pfx string, w, s time.Duration) *Counter {
	return &Counter{
		pool: pool,
		pfx:  pfx,
		w:    w,
		s:    s,
		bkts: int(w/s) + 1,
	}
}

// hash a time to n buckets(n=c.bkts)
func (c *Counter) hash(t int64) int {
	return int(t/int64(c.s)) % c.bkts
}

func (c *Counter) key(id string) string {
	buf := bytes.NewBufferString(c.pfx)
	buf.WriteString(":")
	buf.WriteString(id)
	return buf.String()
}

// increment count in specific bucket
func (c *Counter) inc(id string, bucket int) error {
	conn := c.pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	conn.Send("HINCRBY", c.key(id), strconv.Itoa(bucket), 1)
	conn.Send("HDEL", c.key(id), strconv.Itoa(bucket+1))
	conn.Send("PEXPIRE", c.key(id), int64(c.w/time.Millisecond))
	_, err := conn.Do("EXEC")

	return err
}

// Inc increment id's occurs with current timestamp,
// the count before Counter.w will be cleanup
func (c *Counter) Inc(id string) error {
	now := time.Now().UnixNano()
	bucket := c.hash(now)
	return c.inc(id, bucket)
}

// sum multiple buckets' count, return total
func (c *Counter) count(id string, buckets []int) (int64, error) {
	args := make([]interface{}, len(buckets)+1)
	args[0] = c.key(id)
	for i, v := range buckets {
		args[i+1] = strconv.Itoa(v)
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

// return available buckets
func (c *Counter) buckets(now int) []int {
	rs := make([]int, c.bkts-1)
	for i := 0; i < c.bkts-1; i++ {
		rs[i] = (c.bkts + now - i) % c.bkts
	}
	return rs
}

// Count return total occurs in period of Counter.w
func (c *Counter) Count(id string) (int64, error) {
	now := time.Now().UnixNano()
	buckets := c.buckets(c.hash(now))
	return c.count(id, buckets)
}

// Reset cleanup occurs, set it to zero
func (c *Counter) Reset(id string) error {
	conn := c.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", c.key(id))
	return err
}
