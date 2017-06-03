package stream

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ubclaunchpad/cumulus/message"
)

func TestNewListener(t *testing.T) {
	stream := New(nil)
	resource := map[string]int{
		"hello": 1,
		"world": 500,
	}
	response := &message.Response{
		ID:       uuid.New().String(),
		Error:    message.NewProtocolError(message.InvalidResourceType, "Invalid resource type"),
		Resource: resource,
	}
	lchan := stream.NewListener(response.ID)

	go func(lchan chan *message.Response, response *message.Response) {
		var received *message.Response
		select {
		case received = <-lchan:
			break
		case <-time.After(time.Second * 3):
			t.FailNow()
		}
		if received.ID != response.ID || received.Error != response.Error {
			t.FailNow()
		}
		if received.Resource.(map[string]int)["hello"] != 1 ||
			received.Resource.(map[string]int)["world"] != 500 {
			t.FailNow()
		}
	}(lchan, response)

	go func(lchan chan *message.Response, response *message.Response) {
		lchan <- response
	}(lchan, response)
}

func TestAsyncListeners(t *testing.T) {
	stream := New(nil)
	numlisteners := 10

	testRoutine := func(stream *Stream, numlisteners int) {
		ids := make([]string, numlisteners)
		lchans := make([]chan *message.Response, numlisteners)
		for i := 0; i < numlisteners; i++ {
			ids[i] = uuid.New().String()
			lchans[i] = stream.NewListener(ids[i])
		}
		for i := 0; i < numlisteners; i++ {
			lchan := stream.Listener(ids[i])
			if lchan != lchans[i] {
				t.FailNow()
			}
		}
		for i := 0; i < numlisteners; i++ {
			stream.RemoveListener(ids[i])
			lchan := stream.Listener(ids[i])
			if lchan != nil {
				t.FailNow()
			}
		}
	}

	go testRoutine(stream, numlisteners)
	go testRoutine(stream, numlisteners)
	go testRoutine(stream, numlisteners)
	go testRoutine(stream, numlisteners)
}
