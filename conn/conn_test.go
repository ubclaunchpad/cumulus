package conn

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	_, err := Dial("www.google.ca:80")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
}

func TestListen(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(5)

	handler := func(c net.Conn) {
		defer c.Close()
		buf := make([]byte, 1)
		n, err := c.Read(buf)
		if err != nil {
			t.Fail()
		}
		if n != 1 {
			t.Fail()
		}
		wg.Done()
	}

	go Listen(":8080", handler)
	// Sleep to guarantee that our listener is ready when we start making connections
	time.Sleep(time.Millisecond)

	for i := 0; i < 5; i++ {
		go func() {
			c, err := Dial(":8080")
			if err != nil {
				t.Fail()
				return
			}
			c.Write([]byte{byte(i)})
		}()
	}

	wg.Wait()
	fmt.Println("ending")
}
