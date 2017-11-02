package app

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"io"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
)

/*
https://golang.org/src/crypto/cipher/example_test.go
*/

const nonceSize = 12
const saltSize = 16

// Encrypt encrypts a file with a password
func Encrypt(filePath string, password string) {

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.WithError(err).Fatal("Failed to open file")
	}

	plainText, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.WithError(err).Fatal("Failed to read opened file")
	}

	passwordBytes := []byte(password)

	salt := make([]byte, saltSize)
	_, err = rand.Read(salt)
	if err != nil {
		log.WithError(err).Fatal("Failed to generate salt")
	}

	key := pbkdf2.Key(passwordBytes, salt, 4096, 32, sha512.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.WithError(err).Fatal("Failed to generate AES cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.WithError(err).Fatal("Failed to generate new AES-GCM")
	}

	nonce := make([]byte, nonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		log.WithError(err).Fatal("Failed to generate nonce")
	}

	cipherText := aesgcm.Seal(nil, nonce, plainText, nil)
	cipherText = append(cipherText, salt...)
	cipherText = append(cipherText, nonce...)

	file, err := os.Create(filePath)
	if err != nil {
		log.WithError(err).Fatal("Failed to create new file for cipher text")
	}

	_, err = io.Copy(file, bytes.NewReader(cipherText))
	if err != nil {
		log.WithError(err).Fatal("Failed to copy cipher text to file")
	}
}

// Decrypt decrypts a file with a given password
func Decrypt(filePath string, password string) {

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.WithError(err).Fatal("Failed to open file")
	}

	cipherText, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.WithError(err).Fatal("Failed to read opened file")
	}

	passwordBytes := []byte(password)

	ctLen := len(cipherText)
	nonce := cipherText[ctLen-nonceSize:]
	salt := cipherText[ctLen-nonceSize-saltSize : ctLen-nonceSize]

	key := pbkdf2.Key(passwordBytes, salt, 4096, 32, sha512.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		log.WithError(err).Fatal("Failed to generate AES cipher")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.WithError(err).Fatal("Failed to generate new AES-GCM")
	}

	plainText, err := aesgcm.Open(
		nil,
		nonce,
		cipherText[:ctLen-saltSize-nonceSize],
		nil,
	)
	if err != nil {
		log.WithError(err).Fatal("Failed to generate plain text")
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.WithError(err).Fatal("Failed to create new file for plain text")
	}

	_, err = io.Copy(file, bytes.NewReader(plainText))
	if err != nil {
		log.WithError(err).Fatal("Failed to copy plain text to file")
	}
}
