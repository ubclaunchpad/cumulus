package util

import (
	"math"
	"testing"
)

func TestAppendUint32(t *testing.T) {
	var buf []byte
	buf = AppendUint32(buf, math.MaxUint32)

	if len(buf) != 4 {
		t.Fail()
	}

	for i := 0; i < 4; i++ {
		if buf[i] != 255 {
			t.Fail()
		}
	}

	buf = AppendUint32(buf, math.MaxUint32)

	if len(buf) != 8 {
		t.Fail()
	}

	for i := 0; i < 8; i++ {
		if buf[i] != 255 {
			t.Fail()
		}
	}

	buf = AppendUint32(buf, 0)

	if len(buf) != 12 {
		t.Fail()
	}

	for i := 0; i < 8; i++ {
		if buf[i] != 255 {
			t.Fail()
		}
	}

	for i := 8; i < 12; i++ {
		if buf[i] != 0 {
			t.Fail()
		}
	}
}

func TestAppendUint64(t *testing.T) {
	var buf []byte
	buf = AppendUint64(buf, math.MaxUint64)

	if len(buf) != 8 {
		t.Fail()
	}

	for i := 0; i < 8; i++ {
		if buf[i] != 255 {
			t.Fail()
		}
	}

	buf = AppendUint64(buf, math.MaxUint64)

	if len(buf) != 16 {
		t.Fail()
	}

	for i := 0; i < 16; i++ {
		if buf[i] != 255 {
			t.Fail()
		}
	}

	buf = AppendUint64(buf, 0)

	if len(buf) != 24 {
		t.Fail()
	}

	for i := 0; i < 16; i++ {
		if buf[i] != 255 {
			t.Fail()
		}
	}

	for i := 16; i < 24; i++ {
		if buf[i] != 0 {
			t.Fail()
		}
	}
}
