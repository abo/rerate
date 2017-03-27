rerate 
===========
[![Build Status](https://travis-ci.org/abo/rerate.svg?branch=master)](https://travis-ci.org/abo/rerate)
[![GoDoc](https://godoc.org/github.com/abo/rerate?status.svg)](https://godoc.org/github.com/abo/rerate)
[![Go Report Card](https://goreportcard.com/badge/github.com/abo/rerate)](https://goreportcard.com/report/github.com/abo/rerate)
[![Coverage Status](https://coveralls.io/repos/github/abo/rerate/badge.svg?branch=master)](https://coveralls.io/github/abo/rerate?branch=master)

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
    // redigo buckets
    pool := newRedisPool("localhost:6379", "")
    buckets := rerate.NewRedigoBuckets(pool)

    // OR redis buckets
    // client := redis.NewClient(&redis.Options{
	//	 Addr:     "localhost:6379",
	// 	 Password: "",
	// 	 DB:       0,
	// })
    // buckets := rerate.NewRedisBuckets(client)
    
    // Counter
    counter := rerate.NewCounter(buckets, "rl:test", 10 * time.Minute, 15 * time.Second)
    counter.Inc("click")
    c, err := counter.Count("click")
    
    // Limiter
    limiter := rerate.NewLimiter(buckets, "rl:test", 1 * time.Hour, 15 * time.Minute, 100)
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


Sample - Sparkline
------------------

![](https://github.com/abo/rerate/raw/master/cmd/sparkline/sparkline.png)

```
    cd cmd/sparkline
    npm install webpack -g
    npm install
    webpack && go run main.go
```
Open `http://localhost:8080` in Browser, And then move mouse.


Contributing
------------
WELCOME


License
-------

rerate is available under the [The MIT License (MIT)](https://opensource.org/licenses/MIT).
