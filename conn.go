package pool

import (
	"errors"
)

type PoolConn struct {
	Conn    Conn
	pool    *pool
	useless bool
}

type Conn interface {
	Close() error
}

func (c *PoolConn) Close() error {
	if c.pool == nil || c.Conn == nil {
		return errors.New("pool is nil or conn is nil")
	}

	if c.useless {
		c.Conn.Close()
		return nil
	}
	return c.pool.put(c)
}

func (c *PoolConn) MarkUseless() {
	c.useless = true
}
