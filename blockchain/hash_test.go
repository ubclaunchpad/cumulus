package blockchain

import (
	"testing"
)

func TestParseString(t *testing.T) {
	var h Hash
	if h.ParseString("") != nil {
		t.Fail()
	}

	if h.ParseString("1") != nil {
		t.Fail()
	}

	if h.ParseString("a") != nil {
		t.Fail()
	}

	if h.ParseString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff") != nil {
		t.Fail()
	}

	var emptyHash Hash
	for i := 0; i < HashLen; i++ {
		emptyHash[i] = 0
	}

	if *(h.ParseString("0x")) != emptyHash || *(h.ParseString("0x00")) != emptyHash {
		t.Fail()
	}

	if h.ParseString("0x0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff") != nil {
		t.Fail()
	}

	var one Hash
	one[0] = 1
	for i := 1; i < HashLen; i++ {
		one[i] = 0
	}

	if *(h.ParseString("0x01")) != one {
		t.Fail()
	}

	var maxHash Hash
	for i := 0; i < HashLen; i++ {
		maxHash[i] = 255
	}

	if *(h.ParseString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")) != maxHash {
		t.Fail()
	}

}

func TestCompareTo(t *testing.T) {
	var h1, h2 Hash
	h1.ParseString("0x10000000000008b3a41b85b8b29ad444def299fee21793cd8b9e567eab02cd81")
	h2.ParseString("0x00000000000008a3a41b85b8b29ad444def299fee21793cd8b9e567eab02cd81")

	if !h1.CompareTo(h2, GreaterThan) {
		t.Fail()
	}

	h1.ParseString("0x100000000000000000000000000000000000000000000000000000000000000F")
	h2.ParseString("0xF000000000000000000000000000000000000000000000000000000000000001")

	if !h1.CompareTo(h2, LessThan) {
		t.Fail()
	}

	h1.ParseString("0x0000000000000000000000000000000000000000000000000000000000000000")
	h2.ParseString("0x0000000000000000000000000000000000000000000000000000000000000000")

	if !h1.CompareTo(h2, EqualTo) {
		t.Fail()
	}

	h1.ParseString("0x0000000000000000000000000000000000000000000000000000000000000000")
	h2.ParseString("0x0000000000000000000000000000000000000000000000000000000000000001")

	if !h1.CompareTo(h2, LessThan) {
		t.Fail()
	}

	h1.ParseString("0x0000000000000000000000000000000000000000000000000000000000000001")
	h2.ParseString("0x0000000000000000000000000000000000000000000000000000000000000000")

	if !h1.CompareTo(h2, GreaterThan) {
		t.Fail()
	}

	h1.ParseString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	h2.ParseString("0x0000000000000000000000000000000000000000000000000000000000000000")

	if !h1.CompareTo(h2, GreaterThan) {
		t.Fail()
	}

	h1.ParseString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	h2.ParseString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	if !h1.CompareTo(h2, EqualTo) {
		t.Fail()
	}

	h1.ParseString("0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1")
	h2.ParseString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	if !h1.CompareTo(h2, LessThan) {
		t.Fail()
	}

	h1.ParseString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	h2.ParseString("0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1")

	if !h1.CompareTo(h2, GreaterThan) {
		t.Fail()
	}

	h1.ParseString("0xfffffffffffffffffffffffffffffff11fffffffffffffffffffffffffffffff")
	h2.ParseString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	if !h1.CompareTo(h2, LessThan) {
		t.Fail()
	}
}

func TestHexString(t *testing.T) {
	var maxHash Hash
	for i := 0; i < HashLen; i++ {
		maxHash[i] = 255
	}

	if maxHash.HexString() != "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" {
		t.Fail()
	}

	var emptyHash Hash
	for i := 0; i < HashLen; i++ {
		emptyHash[i] = 0
	}

	if emptyHash.HexString() != "0x0000000000000000000000000000000000000000000000000000000000000000" {
		t.Fail()
	}

	var one Hash
	one[0] = 1
	for i := 1; i < HashLen; i++ {
		one[i] = 0
	}

	if one.HexString() != "0x0000000000000000000000000000000000000000000000000000000000000001" {
		t.Fail()
	}
}
