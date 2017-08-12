package conn

import (
	"bytes"
	"net"
	"time"
)

// TestAddr implementes net.Addr
type TestAddr struct {
	Addr string
}

// Network returns the address of this TestAddr
func (fa TestAddr) Network() string { return fa.Addr }

// Network returns the address of this TestAddr
func (fa TestAddr) String() string { return fa.Addr }

// BufConn implemented net.Conn
type BufConn struct {
	Buf         *bytes.Buffer
	IgnoreWrite bool
	IgnoreRead  bool
	OnRead      func() []byte
	OnWrite     func([]byte)
}

// NewBufConn returns an emty Buffered Connection
func NewBufConn(ignoreWrite bool, ignoreRead bool) *BufConn {
	return &BufConn{
		Buf:         &bytes.Buffer{},
		IgnoreWrite: ignoreWrite,
		IgnoreRead:  ignoreRead,
	}
}

// Read reads bytes from bc into b, and returns the number of bytes read, or
// an error.
func (bc BufConn) Read(b []byte) (int, error) {
	if bc.IgnoreRead {
		return len(b), nil
	}
	if bc.OnRead != nil {
		readBytes := bc.OnRead()
		copy(b, readBytes)
		return len(b), nil
	}
	return bc.Buf.Read(b)
}

// Write appends the contents of p to the buffer, growing the buffer as needed.
// The return value n is the length of p; err is always nil. If the buffer
// becomes too large, Write will panic with ErrTooLarge.
func (bc BufConn) Write(p []byte) (int, error) {
	if bc.IgnoreWrite {
		return len(p), nil
	}
	if bc.OnWrite != nil {
		bc.OnWrite(p)
		return len(p), nil
	}
	return bc.Buf.Write(p)
}

// Close does nothing and returns nil
func (bc BufConn) Close() error { return nil }

// LocalAddr returns the address of the given TestConn so it implementes net.Conn
func (bc BufConn) LocalAddr() net.Addr { return TestAddr{Addr: ""} }

// RemoteAddr return the address of the given TestConn so it implementes net.Conn
func (bc BufConn) RemoteAddr() net.Addr { return TestAddr{Addr: ""} }

// SetDeadline does nothing and returns nil
func (bc BufConn) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline does nothing and returns nil
func (bc BufConn) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline does nothing and returns nil
func (bc BufConn) SetWriteDeadline(t time.Time) error { return nil }
