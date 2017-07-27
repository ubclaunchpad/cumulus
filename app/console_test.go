package app

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunConsoleHasCommands(t *testing.T) {
	a := createNewTestApp()
	s := RunConsole(a)
	expected := []string{
		"address",
		"check",
		"clear",
		"connect",
		"create",
		"exit",
		"help",
		"peers",
	}
	c := s.Cmds()
	found := []string{}
	for i := range c {
		found = append(found, c[i].Name)
	}
	sort.Strings(found)
	assert.Equal(t, expected, found)
}
