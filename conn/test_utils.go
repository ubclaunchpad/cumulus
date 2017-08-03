package conn

import (
	"net"
	"sync"
	"time"
)

// TestConn implements net.Conn
type TestConn struct {
	Addr         net.Addr
	Message      **[]byte
	ReadOnce     bool
	BytesWritten chan []byte
	Lock         *sync.RWMutex
	saveOnWrite  bool
}

// NewTestConn returns a new fake connection with the given settings for testing.
func NewTestConn(addr net.Addr, message **[]byte, readOnce bool, saveOnWrite bool) *TestConn {
	return &TestConn{
		Addr:         addr,
		Message:      message,
		ReadOnce:     readOnce,
		BytesWritten: make(chan []byte),
		Lock:         &sync.RWMutex{},
		saveOnWrite:  saveOnWrite,
	}
}

// Read synchronously copies at most 512 bytes of the contents of this TestConn's
// Message into the given byte slice. If fc.ReadOnce is true it will erase
// fc.Message before returning so it can't be read again.
func (fc TestConn) Read(b []byte) (n int, err error) {
	fc.Lock.RLock()
	defer fc.Lock.RUnlock()

	message := **fc.Message

	var i int
	for i = 0; i < len(message) && i < 512; i++ {
		b[i] = (message)[i]
	}
	if fc.ReadOnce {
		**fc.Message = make([]byte, 0)
	}
	return i, nil
}

// Write pushes the given byte slice into fc.BytesWritten only if fc.saveOnWrite
// is set to true.
func (fc TestConn) Write(b []byte) (n int, err error) {
	if fc.saveOnWrite {
		fc.BytesWritten <- b
	}
	return len(b), nil
}

// Close does nothing and returns nil
func (fc TestConn) Close() error { return nil }

// LocalAddr returns the address of the given TestConn so it implementes net.Conn
func (fc TestConn) LocalAddr() net.Addr { return fc.Addr }

// RemoteAddr return the address of the given TestConn so it implementes net.Conn
func (fc TestConn) RemoteAddr() net.Addr { return fc.Addr }

// SetDeadline does nothing and returns nil
func (fc TestConn) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline does nothing and returns nil
func (fc TestConn) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline does nothing and returns nil
func (fc TestConn) SetWriteDeadline(t time.Time) error { return nil }

// TestAddr implementes net.Addr
type TestAddr struct {
	Addr string
}

// Network returns the address of this TestAddr
func (fa TestAddr) Network() string { return fa.Addr }

// Network returns the address of this TestAddr
func (fa TestAddr) String() string { return fa.Addr }
