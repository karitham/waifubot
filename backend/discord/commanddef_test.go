package discord

import (
	"testing"
)

func TestHashDeterminism(t *testing.T) {
	commands := []CommandDef{
		{Name: "roll", Description: "Roll for a character"},
		{Name: "list", Description: "List characters", Options: []OptionDef{
			{Name: "user", Description: "User", Type: OptionUser},
		}},
	}

	hash1 := Hash(commands)
	hash2 := Hash(commands)

	if hash1 != hash2 {
		t.Errorf("Hash() not deterministic: %s != %s", hash1, hash2)
	}
}

func TestHashOrderInsensitivity(t *testing.T) {
	tests := []struct {
		name      string
		commands1 []CommandDef
		commands2 []CommandDef
	}{
		{
			name: "commands order",
			commands1: []CommandDef{
				{Name: "alpha", Description: "A"},
				{Name: "beta", Description: "B"},
				{Name: "gamma", Description: "C"},
			},
			commands2: []CommandDef{
				{Name: "gamma", Description: "C"},
				{Name: "alpha", Description: "A"},
				{Name: "beta", Description: "B"},
			},
		},
		{
			name: "options order",
			commands1: []CommandDef{
				{
					Name:        "test",
					Description: "Test",
					Options: []OptionDef{
						{Name: "a", Description: "A", Type: OptionString},
						{Name: "b", Description: "B", Type: OptionInt},
					},
				},
			},
			commands2: []CommandDef{
				{
					Name:        "test",
					Description: "Test",
					Options: []OptionDef{
						{Name: "b", Description: "B", Type: OptionInt},
						{Name: "a", Description: "A", Type: OptionString},
					},
				},
			},
		},
		{
			name: "choices order",
			commands1: []CommandDef{
				{
					Name:        "test",
					Description: "Test",
					Options: []OptionDef{
						{
							Name:        "choice",
							Description: "Choice",
							Type:        OptionString,
							Choices:     []ChoiceDef{{Name: "a", Value: "a"}, {Name: "b", Value: "b"}},
						},
					},
				},
			},
			commands2: []CommandDef{
				{
					Name:        "test",
					Description: "Test",
					Options: []OptionDef{
						{
							Name:        "choice",
							Description: "Choice",
							Type:        OptionString,
							Choices:     []ChoiceDef{{Name: "b", Value: "b"}, {Name: "a", Value: "a"}},
						},
					},
				},
			},
		},
		{
			name: "nested options order",
			commands1: []CommandDef{
				{
					Name:        "test",
					Description: "Test",
					Options: []OptionDef{
						{
							Name:        "group",
							Description: "Group",
							Type:        OptionSubcommandGroup,
							Options: []OptionDef{
								{Name: "sub1", Description: "S1", Type: OptionSubcommand},
								{Name: "sub2", Description: "S2", Type: OptionSubcommand},
							},
						},
					},
				},
			},
			commands2: []CommandDef{
				{
					Name:        "test",
					Description: "Test",
					Options: []OptionDef{
						{
							Name:        "group",
							Description: "Group",
							Type:        OptionSubcommandGroup,
							Options: []OptionDef{
								{Name: "sub2", Description: "S2", Type: OptionSubcommand},
								{Name: "sub1", Description: "S1", Type: OptionSubcommand},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := Hash(tt.commands1)
			hash2 := Hash(tt.commands2)
			if hash1 != hash2 {
				t.Errorf("Hash() order sensitive: %s != %s", hash1, hash2)
			}
		})
	}
}

func TestHashDifferentContent(t *testing.T) {
	base := []CommandDef{
		{Name: "roll", Description: "Roll for a character"},
	}

	tests := []struct {
		name      string
		commands1 []CommandDef
		commands2 []CommandDef
		wantSame  bool
	}{
		{
			name:      "identical",
			commands1: base,
			commands2: base,
			wantSame:  true,
		},
		{
			name:      "different name",
			commands1: []CommandDef{{Name: "roll", Description: "Roll for a character"}},
			commands2: []CommandDef{{Name: "roll2", Description: "Roll for a character"}},
			wantSame:  false,
		},
		{
			name:      "different description",
			commands1: []CommandDef{{Name: "roll", Description: "Roll for a character"}},
			commands2: []CommandDef{{Name: "roll", Description: "Different description"}},
			wantSame:  false,
		},
		{
			name:      "with vs without option",
			commands1: []CommandDef{{Name: "roll", Description: "Roll"}},
			commands2: []CommandDef{{Name: "roll", Description: "Roll", Options: []OptionDef{
				{Name: "user", Description: "User", Type: OptionUser},
			}}},
			wantSame: false,
		},
		{
			name: "different option type",
			commands1: []CommandDef{{Name: "roll", Description: "Roll", Options: []OptionDef{
				{Name: "count", Description: "Count", Type: OptionInt},
			}}},
			commands2: []CommandDef{{Name: "roll", Description: "Roll", Options: []OptionDef{
				{Name: "count", Description: "Count", Type: OptionString},
			}}},
			wantSame: false,
		},
		{
			name: "required true vs false",
			commands1: []CommandDef{{Name: "test", Description: "Test", Options: []OptionDef{
				{Name: "id", Description: "ID", Type: OptionInt, Required: true},
			}}},
			commands2: []CommandDef{{Name: "test", Description: "Test", Options: []OptionDef{
				{Name: "id", Description: "ID", Type: OptionInt, Required: false},
			}}},
			wantSame: false,
		},
		{
			name: "autocomplete true vs false",
			commands1: []CommandDef{{Name: "test", Description: "Test", Options: []OptionDef{
				{Name: "id", Description: "ID", Type: OptionInt, Autocomplete: true},
			}}},
			commands2: []CommandDef{{Name: "test", Description: "Test", Options: []OptionDef{
				{Name: "id", Description: "ID", Type: OptionInt, Autocomplete: false},
			}}},
			wantSame: false,
		},
		{
			name: "with vs without choice",
			commands1: []CommandDef{{Name: "roll", Description: "Roll", Options: []OptionDef{
				{Name: "type", Description: "Type", Type: OptionString},
			}}},
			commands2: []CommandDef{{Name: "roll", Description: "Roll", Options: []OptionDef{
				{Name: "type", Description: "Type", Type: OptionString, Choices: []ChoiceDef{
					{Name: "waifu", Value: "waifu"},
				}},
			}}},
			wantSame: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := Hash(tt.commands1)
			hash2 := Hash(tt.commands2)
			if tt.wantSame && hash1 != hash2 {
				t.Errorf("Hash() = %s vs %s, want same", hash1, hash2)
			}
			if !tt.wantSame && hash1 == hash2 {
				t.Errorf("Hash() = %s, unexpectedly same for different content", hash1)
			}
		})
	}
}

func TestHashDoesNotModifyInput(t *testing.T) {
	commands := []CommandDef{
		{Name: "zeta", Description: "Z"},
		{Name: "alpha", Description: "A"},
	}
	original := make([]CommandDef, len(commands))
	copy(original, commands)

	Hash(commands)

	for i := range commands {
		if commands[i].Name != original[i].Name {
			t.Errorf("Hash() modified input: commands[%d].Name = %s, want %s", i, commands[i].Name, original[i].Name)
		}
	}
}

func TestToCordeSlice(t *testing.T) {
	commands := []CommandDef{
		{Name: "roll", Description: "Roll"},
		{Name: "list", Description: "List"},
	}

	got := ToCorde(commands)
	if len(got) != 2 {
		t.Errorf("ToCorde() returned %d commands, want 2", len(got))
	}
	for i := range commands {
		if got[i] == nil {
			t.Errorf("ToCorde()[%d] is nil", i)
		}
	}
}
