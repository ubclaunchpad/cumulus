package moj

import (
	"math"
	"math/big"
	"testing"

	"gopkg.in/kyokomi/emoji.v1"

	"github.com/stretchr/testify/assert"
)

// encoder takes a single rune and encodes an emoji.
type encoder interface {
	encode(r rune) (string, error)
}

type intEncoder struct{}
type hexEncoder struct{}
type bigEncoder struct{}

func (e intEncoder) encode(r rune) (string, error) {
	switch {
	case r >= 'A':
		return EncodeInt(int(r) - 55)
	default:
		return EncodeInt(int(r) - 48)
	}
}

func (e hexEncoder) encode(r rune) (string, error) {
	switch {
	case r >= 'A':
		return EncodeHex(string(r))
	default:
		return EncodeHex(string(r))
	}
}

func (e bigEncoder) encode(r rune) (string, error) {
	switch {
	case r >= 'A':
		return EncodeBigInt(big.NewInt(int64(r - 55)))
	default:
		return EncodeBigInt(big.NewInt(int64(r - 48)))
	}
}

func BasicTestRun(e encoder, t *testing.T) {
	var result string
	var err error
	for r, moj := range emojiRuneMap {
		result, err = e.encode(r)
		assert.Nil(t, err)
		// Encoding converts to the unicode symbol.
		// Use the codemap to make sure we got the
		// right unicode symbol.
		assert.Equal(t, result, emoji.CodeMap()[moj])
	}
}

func TestEncodeIntBasic(t *testing.T) {
	BasicTestRun(intEncoder{}, t)
}

func TestEncodeHexBasic(t *testing.T) {
	BasicTestRun(hexEncoder{}, t)
}

func TestEncodeBigBasic(t *testing.T) {
	BasicTestRun(bigEncoder{}, t)
}

func TestEncodeIntAdvanced(t *testing.T) {

	// 7f
	result, _ := EncodeInt(math.MaxInt8)
	expected := "👏 🍖"
	assert.Equal(t, result, expected)

	// 7fff
	result, _ = EncodeInt(math.MaxInt16)
	expected = "👏 🍖 🍖 🍖"
	assert.Equal(t, result, expected)

	// 7fff_ffff
	result, _ = EncodeInt(math.MaxInt32)
	expected = "👏 🍖 🍖 🍖 🍖 🍖 🍖 🍖"
	assert.Equal(t, result, expected)
}

func TestEncodeHexAdvanced(t *testing.T) {

	result, _ := EncodeHex("abcdef1234567890")
	println(result)
	expected := "🙌 🤘 🕑 🚁 ⚓️ 🍖 ☝️ ✌️ 🌵 👍 ✋ 🏃 👏 🍄 🌷 ✊"
	assert.Equal(t, result, expected)

	result, _ = EncodeHex("badf00d")
	expected = "🤘 🙌 🚁 🍖 ✊ ✊ 🚁"
	assert.Equal(t, result, expected)

	result, _ = EncodeHex("abadd00d")
	expected = "🙌 🤘 🙌 🚁 🚁 ✊ ✊ 🚁"
	assert.Equal(t, result, expected)

	result, _ = EncodeHex("deadbeef")
	expected = "🚁 ⚓️ 🙌 🚁 🤘 ⚓️ ⚓️ 🍖"
	assert.Equal(t, result, expected)

	result, _ = EncodeHex("70ffee")
	expected = "👏 ✊ 🍖 🍖 ⚓️ ⚓️"
	assert.Equal(t, result, expected)
}
