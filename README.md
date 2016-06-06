## Pool
Pool, written by Go, is a thread safe connection pool. It can be used to manage any connections that implenment pool.Conn interface.

## Import

```go
import "github.com/gowithrain/pool"

```

## pool.Conn

pool.Conn is a interface and include Close methond.

```go
type Conn interface {
	Close() error
}
```

	
## Example

```go
// in main func
	
func main() {
	newf := func() (pool.Conn, error) {
		return net.DialTimeout("tcp", "127.0.0.1:8000", time.Duration(1000)*time.Millisecond)
	}
	
	p, err := pool.New(5, 20, newf)
	
	conn, err := p.Get()
	if err != nil {
		// error
		return
	}
	defer conn.Close()
	
	c, ok := conn.Conn.(net.Conn)
	if !ok {
		// error
		return
	}
	
	buf := make([]byte, 10)
	n, err := c.Read(buf)
	if err != nil {
		conn.MarkUseless()
		return
	}
	
	/// do something
	return
}
```