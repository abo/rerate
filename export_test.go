package rerate

func Buckets(c *Counter, from int) []int {
	return c.buckets(from)
}

func Hash(c *Counter, t int64) int {
	return c.hash(t)
}

func IncBucket(c *Counter, id string, bucket int) error {
	return c.inc(id, bucket)
}

func HistogramFrom(c *Counter, id string, from int) ([]int64, error) {
	return c.histogram(id, from)
}
