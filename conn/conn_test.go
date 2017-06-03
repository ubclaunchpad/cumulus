package conn

import (
	"fmt"
	"net"
	"testing"
)

func TestConnect(t *testing.T) {
	_, err := Dial("www.google.ca:80")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
}

func TestListen(t *testing.T) {
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
	}

	go Listen(":8080", handler)

	for i := 0; i < 5; i++ {
		go func() {
			c, err := Dial(":8080")
			if err != nil {
				t.Fail()
			}
			c.Write([]byte{byte(i)})
		}()
	}
}
