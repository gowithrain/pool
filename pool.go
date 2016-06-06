package pool

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrClosed = errors.New("pool is closed")
)

type Pool interface {
	Get() (*PoolConn, error)
	Close()
	Len() int
}

type pool struct {
	mu    sync.Mutex
	conns chan *PoolConn
	newf  NewF
}

type NewF func() (Conn, error)

func New(initConnNum, maxIdle int, newf NewF) (Pool, error) {
	if initConnNum < 0 || maxIdle <= 0 || initConnNum > maxIdle {
		return nil, errors.New("invalid capacity settings")
	}

	p := &pool{
		conns: make(chan *PoolConn, maxIdle),
		newf:  newf,
	}

	for i := 0; i < initConnNum; i++ {
		c, err := newf()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("newf is not able to fill the pool:%s", err)
		}
		conn := &PoolConn{Conn: c, pool: p}
		p.conns <- conn
	}
	return p, nil
}

func (p *pool) Get() (*PoolConn, error) {
	p.mu.Lock()
	if p.conns == nil {
		return nil, ErrClosed
	}

	select {
	case conn := <-p.conns:
		p.mu.Unlock()
		if conn == nil {
			return nil, ErrClosed
		}
		return conn, nil
	default:
		p.mu.Unlock()
		c, err := p.newf()
		if err != nil {
			return nil, err
		}
		conn := &PoolConn{Conn: c, pool: p}
		return conn, nil
	}
}

func (p *pool) put(conn *PoolConn) error {
	if conn == nil {
		return errors.New("connection is nil, rejecting")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conns == nil {
		conn.Conn.Close()
		return ErrClosed
	}

	select {
	case p.conns <- conn:
		return nil
	default:
		conn.Conn.Close()
		return errors.New("pool is full")
	}
}

func (p *pool) Close() {
	p.mu.Lock()
	conns := p.conns
	p.conns = nil
	p.newf = nil
	p.mu.Unlock()

	if conns == nil {
		return
	}
	close(conns)
	for conn := range conns {
		conn.Conn.Close()
	}
}

func (p *pool) Len() int {
	return len(p.conns)
}
