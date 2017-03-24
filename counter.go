package rerate

import (
	"fmt"
	"time"
)

// Counter count total occurs during a period,
// it will store occurs during every time slice interval: (now ~ now - interval), (now - interval ~ now - 2*interval)...
type Counter struct {
	pfx      string
	buckets  Buckets
	period   time.Duration
	interval time.Duration
}

// NewCounter create a new Counter
func NewCounter(newBuckets BucketsFactory, prefix string, period, interval time.Duration) *Counter {
	return &Counter{
		buckets:  newBuckets(int64(period/interval), period),
		pfx:      prefix,
		period:   period,
		interval: interval,
	}
}

// hash a time to n buckets(n=c.bkts)
func (c *Counter) hash(t time.Time) int64 {
	return t.UnixNano() / int64(c.interval)
}

func (c *Counter) key(id string) string {
	return fmt.Sprintf("%s:%s", c.pfx, id)
}

func (c *Counter) incAt(id string, t time.Time) error {
	bucketID := c.hash(t)
	if err := c.buckets.Inc(c.key(id), bucketID); err != nil {
		return err
	}
	return nil
}

// Inc increment id's occurs with current timestamp,
// the count before period will be cleanup
func (c *Counter) Inc(id string) error {
	return c.incAt(id, time.Now())
}

func (c *Counter) histogramAt(id string, t time.Time) ([]int64, error) {
	from := c.hash(t)
	size := int(c.period / c.interval)
	bucketIDs := make([]int64, size)
	for i := 0; i < size; i++ {
		bucketIDs[i] = from - int64(i)
	}

	return c.buckets.Get(c.key(id), bucketIDs...)
}

// Histogram return count histogram in recent period, order by time desc
func (c *Counter) Histogram(id string) ([]int64, error) {
	return c.histogramAt(id, time.Now())
}

func (c *Counter) countAt(id string, t time.Time) (int64, error) {
	h, err := c.histogramAt(id, t)
	if err != nil {
		return 0, err
	}

	total := int64(0)
	for _, v := range h {
		total += v
	}
	return total, nil
}

// Count return total occurs in recent period
func (c *Counter) Count(id string) (int64, error) {
	return c.countAt(id, time.Now())
}

// Reset cleanup occurs, set it to zero
func (c *Counter) Reset(id string) error {
	return c.buckets.Del(c.key(id))
}
