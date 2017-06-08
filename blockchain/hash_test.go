package blockchain

import "testing"

func TestCompareTo(t *testing.T) {
	h1 := HexStringToHash("00000000000008b3a41b85b8b29ad444def299fee21793cd8b9e567eab02cd81")
	h2 := HexStringToHash("00000000000008a3a41b85b8b29ad444def299fee21793cd8b9e567eab02cd81")

	if !CompareTo(h1, h2, GreaterThan) {
		t.Fail()
	}

	h1 = HexStringToHash("100000000000000000000000000000000000000000000000000000000000000F")
	h2 = HexStringToHash("F000000000000000000000000000000000000000000000000000000000000001")

	if !CompareTo(h1, h2, LessThan) {
		t.Fail()
	}

	h1 = HexStringToHash("0000000000000000000000000000000000000000000000000000000000000000")
	h2 = HexStringToHash("0000000000000000000000000000000000000000000000000000000000000000")

	if !CompareTo(h1, h2, EqualTo) {
		t.Fail()
	}

	h1 = HexStringToHash("0000000000000000000000000000000000000000000000000000000000000000")
	h2 = HexStringToHash("0000000000000000000000000000000000000000000000000000000000000001")

	if !CompareTo(h1, h2, LessThan) {
		t.Fail()
	}

	h1 = HexStringToHash("0000000000000000000000000000000000000000000000000000000000000001")
	h2 = HexStringToHash("0000000000000000000000000000000000000000000000000000000000000000")

	if !CompareTo(h1, h2, GreaterThan) {
		t.Fail()
	}

	h1 = HexStringToHash("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	h2 = HexStringToHash("0000000000000000000000000000000000000000000000000000000000000000")

	if !CompareTo(h1, h2, GreaterThan) {
		t.Fail()
	}

	h1 = HexStringToHash("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	h2 = HexStringToHash("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	if !CompareTo(h1, h2, EqualTo) {
		t.Fail()
	}

	h1 = HexStringToHash("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1")
	h2 = HexStringToHash("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	if !CompareTo(h1, h2, LessThan) {
		t.Fail()
	}

	h1 = HexStringToHash("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	h2 = HexStringToHash("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1")

	if !CompareTo(h1, h2, GreaterThan) {
		t.Fail()
	}

	h1 = HexStringToHash("fffffffffffffffffffffffffffffff11fffffffffffffffffffffffffffffff")
	h2 = HexStringToHash("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	if !CompareTo(h1, h2, LessThan) {
		t.Fail()
	}
}
