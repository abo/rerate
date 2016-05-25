package ratelimiter

import (
	"github.com/garyburd/redigo/redis"
)

// Pool maintains a pool of connections. The application calls the Get method
// to get a connection from the pool and the connection's Close method to
// return the connection's resources to the pool.
type Pool interface {
	Get() redis.Conn
}
