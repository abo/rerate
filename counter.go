package rerate

import (
	"bytes"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Counter count total occurs during a period,
// it will store occurs during every time slice interval: (now ~ now - intervl), (now - intervl ~ now - 2*intervl)...
type Counter struct {
	pool     Pool
	pfx      string
	period   time.Duration
	interval time.Duration
	bkts     int
}

// NewCounter create a new Counter
func NewCounter(pool Pool, prefix string, period, interval time.Duration) *Counter {
	return &Counter{
		pool:     pool,
		pfx:      prefix,
		period:   period,
		interval: interval,
		bkts:     int(period/interval) * 2,
	}
}

// hash a time to n buckets(n=c.bkts)
func (c *Counter) hash(t int64) int {
	return int(t/int64(c.interval)) % c.bkts
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

	args := make([]interface{}, (c.bkts/2)+1)
	args[0] = c.key(id)
	for i := 0; i < c.bkts/2; i++ {
		args[i+1] = (bucket + i + 1) % c.bkts
	}

	conn.Send("MULTI")
	conn.Send("HINCRBY", c.key(id), strconv.Itoa(bucket), 1)
	conn.Send("HDEL", args...)
	conn.Send("PEXPIRE", c.key(id), int64(c.period/time.Millisecond))
	_, err := conn.Do("EXEC")

	return err
}

// Inc increment id's occurs with current timestamp,
// the count before period will be cleanup
func (c *Counter) Inc(id string) error {
	now := time.Now().UnixNano()
	bucket := c.hash(now)
	return c.inc(id, bucket)
}

// return available buckets
func (c *Counter) buckets(from int) []int {
	len := c.bkts / 2
	rs := make([]int, len)
	for i := 0; i < len; i++ {
		rs[i] = (c.bkts + from - i) % c.bkts
	}
	return rs
}

func (c *Counter) histogram(id string, from int) ([]int64, error) {
	buckets := c.buckets(from)
	args := make([]interface{}, len(buckets)+1)
	args[0] = c.key(id)
	for i, v := range buckets {
		args[i+1] = v
	}

	conn := c.pool.Get()
	defer conn.Close()

	vals, err := redis.Strings(conn.Do("HMGET", args...))
	if err != nil {
		return []int64{}, err
	}

	ret := make([]int64, len(buckets))
	for i, val := range vals {
		if v, e := strconv.ParseInt(val, 10, 64); e == nil {
			ret[i] = v
		} else {
			ret[i] = 0
		}
	}
	return ret, nil
}

// Histogram return count histogram in recent period, order by time desc
func (c *Counter) Histogram(id string) ([]int64, error) {
	now := time.Now().UnixNano()
	from := c.hash(now)

	return c.histogram(id, from)
}

// Count return total occurs in recent period
func (c *Counter) Count(id string) (int64, error) {
	h, err := c.Histogram(id)
	if err != nil {
		return 0, err
	}

	total := int64(0)
	for _, v := range h {
		total += v
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
