package rerate_test

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	. "github.com/abo/rerate"
)

var pool Pool

func randkey() string {
	return strconv.Itoa(rand.Int())
}

func init() {
	pool = newRedisPool("localhost:6379", "")
}

func TestBuckets(t *testing.T) {
	testcases := map[int][]int{
		1:  {1, 0, 19, 18, 17, 16, 15, 14, 13, 12},
		0:  {0, 19, 18, 17, 16, 15, 14, 13, 12, 11},
		10: {10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
	}

	counter := NewCounter(pool, "rerate:test:counter:buckets", 10*time.Second, time.Second)
	for input, expect := range testcases {
		buckets := Buckets(counter, input)
		if !reflect.DeepEqual(buckets, expect) {
			t.Fatal("expect ", expect, ",but ", buckets)
		}
	}
}

func TestHash(t *testing.T) {
	// hash(a+s) = hash(a)+1
	counter := NewCounter(pool, "rerate:test:counter:hash", time.Minute, time.Second)
	l := int(time.Minute / time.Second)

	testcases := []int64{time.Now().UnixNano(),
		time.Now().UnixNano() - int64(rand.Intn(100))*int64(time.Second),
		time.Now().UnixNano() - int64(rand.Intn(100))*int64(time.Second),
		time.Now().UnixNano() - int64(rand.Intn(100))*int64(time.Second),
		time.Now().UnixNano() - int64(rand.Intn(100))*int64(time.Second)}

	for _, input := range testcases {
		next := input + int64(time.Second)
		b := Hash(counter, input)
		nb := Hash(counter, next)
		if b < 0 || b > 2*l {
			t.Fatal("out of range ", input)
		}
		if nb < 0 || nb > 2*l {
			t.Fatal("out of range ", input)
		}
		if b+1 != nb && b+1-(2*l) != nb {
			t.Fatal("input ", input)
		}

		np := Hash(counter, input+int64(time.Minute))
		if (np-b) != l && (b-np) != l {
			t.Fatal("input", input, " incorrect next period")
		}
	}

}

func TestHistogram(t *testing.T) {
	counter := NewCounter(pool, "rerate:test:counter:count", 4000*time.Millisecond, 400*time.Millisecond)
	id := randkey()
	counter.Reset(id)

	zero := []int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if b, err := counter.Histogram(id); err != nil || !reflect.DeepEqual(b, zero) {
		t.Fatal("expect all zero without err, actual", b, err)
	}

	for i := 0; i <= 10; i++ {
		for j := 0; j < i; j++ {
			IncBucket(counter, id, i)
		}
	} //[]int64{0,1,2,3,4,5,6,7,8,9,10,0,0,0,0,0,0,0,0,0}

	for i := 0; i < 20; i++ {
		for j := len(zero) - 1; j > 0; j-- {
			zero[j] = zero[j-1]
		}
		if i <= 10 {
			zero[0] = int64(i)
		} else {
			zero[0] = 0
		}

		assertHist(t, counter, id, i, zero)
	}
}

func TestCounter(t *testing.T) {
	counter := NewCounter(pool, "rerate:test:counter:counter", time.Minute, time.Second)
	ip1, ip2 := randkey(), randkey()

	if err := counter.Reset(ip1); err != nil {
		t.Fatal("can not reset counter", err)
	}
	if err := counter.Reset(ip2); err != nil {
		t.Fatal("can not reset counter", err)
	}

	assertCount(t, counter, ip1, 0)

	for i := 0; i < 10; i++ {
		counter.Inc(ip1)
		assertCount(t, counter, ip1, int64(i+1))
		assertCount(t, counter, ip2, 0)
	}
}

func assertCount(t *testing.T, c *Counter, k string, expect int64) {
	if count, err := c.Count(k); err != nil || count != expect {
		t.Fatal("should be ", expect, " without error, actual ", count, err)
	}
}

func assertHist(t *testing.T, c *Counter, k string, from int, expect []int64) {
	if b, err := HistogramFrom(c, k, from); err != nil || !reflect.DeepEqual(b, expect) {
		t.Fatal("expect ", expect, " without err, actual", b, err)
	}
}
