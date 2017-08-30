package app

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunConsoleHasCommands(t *testing.T) {
	a := newTestApp()
	s := RunConsole(a)
	expected := []string{
		"address",
		"check",
		"clear",
		"connect",
		"create",
		"exit",
		"help",
		"miner",
		"peers",
		"user",
	}
	c := s.Cmds()
	found := []string{}
	for i := range c {
		found = append(found, c[i].Name)
	}
	sort.Strings(found)
	assert.Equal(t, expected, found)
}
