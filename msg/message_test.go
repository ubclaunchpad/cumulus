package msg

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
)

func TestRequest(t *testing.T) {
	id := uuid.New().String()
	r := Request{
		ID:           id,
		ResourceType: ResourceTransaction,
	}

	var buf bytes.Buffer

	err := r.Write(&buf)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	payload, err := Read(&buf)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	req, ok := payload.(*Request)
	if !ok || req.ID != id {
		t.Fail()
	}
}

func TestResponse(t *testing.T) {
	id := uuid.New().String()
	r := Response{
		ID:       id,
		Resource: "resource",
	}

	var buf bytes.Buffer

	err := r.Write(&buf)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	payload, err := Read(&buf)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	res, ok := payload.(*Response)
	if !ok || res.ID != id {
		t.Fail()
	}

	resource, ok := res.Resource.(string)
	if !ok {
		t.Fail()
	}

	if resource != "resource" {
		t.Fail()
	}
}

func TestPush(t *testing.T) {
	p := Push{
		ResourceType: ResourceTransaction,
		Resource:     "transaction",
	}

	var buf bytes.Buffer

	err := p.Write(&buf)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	payload, err := Read(&buf)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	push, ok := payload.(*Push)
	if !ok {
		t.Fail()
	}

	resource, ok := push.Resource.(string)
	if !ok {
		t.Fail()
	}

	if resource != "transaction" {
		t.Fail()
	}
}

func TestError(t *testing.T) {
	msg := "Not Implemented"
	err := NewProtocolError(NotImplemented, msg)

	if err.Error() != msg {
		t.Fail()
	}
}
