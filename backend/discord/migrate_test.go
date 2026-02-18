package discord

import (
	"testing"
)

func TestCommandDefinitionsHash(t *testing.T) {
	hash := Hash(commandDefinitions)
	if len(hash) != 64 {
		t.Errorf("expected 64-char hash, got %d", len(hash))
	}
}

func TestCommandDefinitionsToCorde(t *testing.T) {
	cmds := ToCorde(commandDefinitions)
	if len(cmds) != len(commandDefinitions) {
		t.Errorf("expected %d commands, got %d", len(commandDefinitions), len(cmds))
	}
}
