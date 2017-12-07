package app_test

import (
	"testing"

	"github.com/ubclaunchpad/cumulus/app"
)

func TestInvalidPassword(t *testing.T) {
	plainText := "Hello, world!"
	plainTextBytes := []byte(plainText)

	cipherText, _ := app.Encrypt(plainTextBytes, "correctPassword")
	plainTextNew, err := app.Decrypt(cipherText, "incorrectPassword")

	if !app.InvalidPassword(err) {
		t.Fail()
	}

	if plainTextNew != nil {
		t.Fail()
	}

	plainTextNew, err = app.Decrypt(cipherText, "correctPassword")

	if string(plainTextNew[:]) != plainText {
		t.Fail()
	}
}
