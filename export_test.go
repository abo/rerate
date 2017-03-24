package rerate

import "time"

// func BucketIdsExp(c *Counter, from int) []int {
// 	return c.bucketIds(from)
// }

func HashExp(c *Counter, t time.Time) int64 {
	return c.hash(t)
}

func IncAtExp(c *Counter, id string, t time.Time) error {
	return c.incAt(id, t)
}

func HistogramAtExp(c *Counter, id string, t time.Time) ([]int64, error) {
	return c.histogramAt(id, t)
}

func RemainingAtExp(l *Limiter, id string, t time.Time) (int64, error) {
	return l.remainingAt(id, t)
}

func ExceededAtExp(l *Limiter, id string, t time.Time) (bool, error) {
	return l.exceededAt(id, t)
}
