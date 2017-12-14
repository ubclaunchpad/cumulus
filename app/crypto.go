package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

/*
https://golang.org/src/crypto/cipher/example_test.go
*/

// Encryption utility constants
const (
	nonceSize = 12
	saltSize  = 16
)

// Password complexity constants
const (
	// minPasswordLen represents the min password length in characters
	minPasswordLen = 10
	// maxPasswordLen = 128 represents the max password length in characters
	maxPasswordLen = 128
)

// Encrypt encrypts cipherText with a given password
func Encrypt(plainText []byte, password string) ([]byte, error) {

	passwordBytes := []byte(password)

	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	key := pbkdf2.Key(passwordBytes, salt, 4096, 32, sha512.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, nonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	cipherText := aesgcm.Seal(nil, nonce, plainText, nil)
	cipherText = append(cipherText, salt...)
	cipherText = append(cipherText, nonce...)

	return cipherText, nil
}

// Decrypt decrypts cipherText with a given password
func Decrypt(cipherText []byte, password string) ([]byte, error) {

	passwordBytes := []byte(password)

	ctLen := len(cipherText)
	nonce := cipherText[ctLen-nonceSize:]
	salt := cipherText[ctLen-nonceSize-saltSize : ctLen-nonceSize]

	key := pbkdf2.Key(passwordBytes, salt, 4096, 32, sha512.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainText, err := aesgcm.Open(
		nil,
		nonce,
		cipherText[:ctLen-saltSize-nonceSize],
		nil,
	)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

// VerifyPasswordComplexity verifies password complexity
func VerifyPasswordComplexity(password string) bool {
	return (len(password) >= minPasswordLen) && (len(password) <= maxPasswordLen)
}

// InvalidPassword is returned from Decrypt if an invalid password is used to
// decrypt the ciphertext
func InvalidPassword(err error) bool {
	if err == nil {
		return false
	}
	return strings.Compare(err.Error(), "cipher: message authentication failed") == 0
}
