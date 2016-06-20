rerate 
===========
[![Build Status](https://api.travis-ci.org/abo/rerate.svg)](https://travis-ci.org/abo/rerate)
[![GoDoc](https://godoc.org/github.com/abo/rerate?status.svg)](https://godoc.org/github.com/abo/rerate)
[![Go Report Card](https://goreportcard.com/badge/github.com/abo/rerate)](https://goreportcard.com/report/github.com/abo/rerate)
[![Coverage](http://gocover.io/_badge/github.com/abo/rerate)](https://gocover.io/github.com/abo/rerate)

rerate is a redis-based ratecounter and ratelimiter

* Dead simple api
* With redis as backend, multiple rate counters/limiters can work as a cluster
* Count/Limit requests any period, 2 day, 1 hour, 5 minute or 2 second, it's up to you
* Recording requests as a histotram, which can be used to visualize or monitor
* Limit requests from single ip, userid, applicationid, or any other unique identifier


Tutorial
--------
```
package main

import (
    "github.com/abo/rerate"
)

...

func main() {
    pool := newRedisPool("localhost:6379", "")
    
    // Counter
    counter := rerate.NewCounter(pool, "rl:test", 10 * time.Minute, 15 * time.Second)
    counter.Inc("click")
    c, err := counter.Count("click")
    
    // Limiter
    limiter := rerate.NewLimiter(pool, "rl:test", 1 * time.Hour, 15 * time.Minute, 100)
    limiter.Inc("114.255.86.200")
    rem, err := limiter.Remaining("114.255.86.200")
    exceed, err := limiter.Exceeded("114.255.86.200")
}
```


Installation
------------

Install rerate using the "go get" command:

    go get github.com/abo/rerate

Documentation
-------------

- [API Reference](http://godoc.org/github.com/abo/rerate)
- [Wiki](https://github.com/abo/rerate/wiki)


Contributing
------------
WELCOME


License
-------

rerate is available under the [The MIT License (MIT)](https://opensource.org/licenses/MIT).
