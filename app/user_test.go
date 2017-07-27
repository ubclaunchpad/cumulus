package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentUser(t *testing.T) {
	assert.NotNil(t, getCurrentUser())
}
