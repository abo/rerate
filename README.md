ratelimiter 
===========
[![Build Status](https://travis-ci.org/abo/ratelimiter.svg)](https://travis-ci.org/abo/ratelimiter)
[![GoDoc](https://godoc.org/github.com/abo/ratelimiter?status.svg)](https://godoc.org/github.com/abo/ratelimiter)

ratelimiter is a redis-based ratecounter and ratelimiter

* [Counter](https://godoc.org/github.com/abo/ratelimiter#Counter) - redis-based counter
* [Limiter](https://godoc.org/github.com/abo/ratelimiter#Limiter) - redis-based limiter

Tutorial
--------
```
package main

import (
    "github.com/abo/ratelimiter"
)

...

func main() {
    pool := newRedisPool("localhost:6379", "")
    
    // Counter
    counter := ratelimiter.NewCounter(pool, "rl:test", 10 * time.Minute, 15 * time.Second)
    counter.Inc("click")
    c, err := counter.Count("click")
    
    // Limiter
    limiter := ratelimiter.NewLimiter(pool, "rl:test", 1 * time.Hour, 15 * time.Minute, 100)
    limiter.Inc("114.255.86.200")
    rem, err := limiter.Remaining("114.255.86.200")
    exceed, err := limiter.Exceeded("114.255.86.200")
}
```


Installation
------------

Install ratelimiter using the "go get" command:

    go get github.com/abo/ratelimiter

Documentation
-------------

- [API Reference](http://godoc.org/github.com/abo/ratelimiter)
- [Wiki](https://github.com/abo/ratelimiter/wiki)


Contributing
------------
WELCOME


License
-------

ratelimiter is available under the [The MIT License (MIT)](https://opensource.org/licenses/MIT).
