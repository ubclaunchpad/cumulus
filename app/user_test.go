package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentUser(t *testing.T) {
	assert.NotNil(t, NewUser())
}

func TestPublic(t *testing.T) {
	user := NewUser()
	assert.NotNil(t, user.Public)
}

func TestSaveAndLoad(t *testing.T) {
	user1 := NewUser()
	assert.Nil(t, user1.Save("userTestFile.json"))
	user2, err := Load("userTestFile.json")
	assert.Nil(t, err)
	assert.Equal(t, user1, user2)
	assert.Nil(t, os.Remove("userTestFile.json"))
}
