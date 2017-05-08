package blockchain

import (
	"fmt"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	w, err := NewWallet()
	if err != nil {
		t.Fail()
	}

	unmarshalled := Unmarshal(w.Marshal())
	fmt.Println("w", w.X, w.Y)
	fmt.Println("u", unmarshalled.X, unmarshalled.Y)
	if !w.Equals(unmarshalled) {
		t.Fatal("Unmarshal produced a different wallet")
	}
}
