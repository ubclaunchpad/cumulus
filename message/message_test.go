package message

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

	out, err := Read(&buf)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	if out.Type() != MessageRequest {
		t.Fail()
	}

	outReq, ok := out.(*Request)
	if !ok {
		t.Fail()
	}

	if outReq.ID != id {
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

	out, err := Read(&buf)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	if out.Type() != MessageResponse {
		t.Fail()
	}

	outRes, ok := out.(*Response)
	if !ok {
		t.Fail()
	}

	if outRes.ID != id {
		t.Fail()
	}

	res, ok := outRes.Resource.(string)
	if !ok {
		t.Fail()
	}

	if res != "resource" {
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

	out, err := Read(&buf)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}

	if out.Type() != MessagePush {
		t.Fail()
	}

	outPush, ok := out.(*Push)
	if !ok {
		t.Fail()
	}

	res, ok := outPush.Resource.(string)
	if !ok {
		t.Fail()
	}

	if res != "transaction" {
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
