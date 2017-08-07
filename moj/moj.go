package moj

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"gopkg.in/kyokomi/emoji.v1"
)

// packageBase constants define the base number system underlying the package.
const packageBase = 16
const packageBaseStringFmt = "%X"

// EncodeInt encodes an integer as a string of emojis.
func EncodeInt(i int) (string, error) {
	return EncodeHex(fmt.Sprintf(packageBaseStringFmt, i))
}

// EncodeBigInt encodes a big integer as a string of emojis.
func EncodeBigInt(i *big.Int) (string, error) {
	return EncodeHex(i.Text(packageBase))
}

// EncodeHex encodes a hex string as a string of emojis.
func EncodeHex(h string) (string, error) {
	var emojiString string
	for _, char := range h {
		switch {
		// Handle numeric rune.
		case char >= '0' && char <= '9':
			emojiString = emojiString + emojiRuneMap[char]
		// Handle uppercase rune.
		case char >= 'A' && char <= 'F':
			emojiString = emojiString + emojiRuneMap[char]
		// Handle lower case rune (map to uppercase).
		case char >= 'a' && char <= 'f':
			emojiString = emojiString + emojiRuneMap[char-32]
		// Other characters raise error.
		default:
			return emojiString, errors.New("hex string is malformed")
		}
	}

	// Remove white space appended by emoji package :/.
	return strings.TrimSpace(emoji.Sprint(emojiString)), nil
}
