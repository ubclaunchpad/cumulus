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

func TestPasswordComplexity(t *testing.T) {
	test := "12345679"
	if app.VerifyPasswordComplexity(test) {
		t.Fail()
	}

	test = "12345678910"
	if !app.VerifyPasswordComplexity(test) {
		t.Fail()
	}

	test = ""
	for i := 0; i < 128; i++ {
		test += "1"
	}
	if !app.VerifyPasswordComplexity(test) {
		t.Fail()
	}

	// Test 128 character length
	test = ""
	for i := 0; i < 129; i++ {
		test += "1"
	}
	if app.VerifyPasswordComplexity(test) {
		t.Fail()
	}
}
