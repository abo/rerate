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

func (c *Counter) bucket() int {
	now := time.Now().UnixNano()
	return int(now/int64(c.s)) % c.bkts
}

func (c *Counter) key(id string) string {
	buf := bytes.NewBufferString(c.pfx)
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
	conn.Send("PEXPIRE", c.key(id), c.w/time.Millisecond)
	_, err := conn.Do("EXEC")

	return err
}

// Count return total occurs in duration(buckets * precision)
func (c *Counter) Count(id string) (int64, error) {
	bucket := c.bucket()
	args := make([]interface{}, c.bkts)
	args[0] = c.key(id)
	for i := 0; i < c.bkts-1; i++ {
		args[i+1] = strconv.Itoa((c.bkts + bucket - i) % c.bkts)
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
