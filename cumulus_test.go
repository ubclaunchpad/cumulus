package cumulus

import (
	"testing"
)

func TestCumulus(t *testing.T) {
	if !dummy() {
		t.Fail()
	}
}
