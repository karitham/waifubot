package discord

import (
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		commands []CommandDef
		wantLen  int
	}{
		{
			name:     "empty commands",
			commands: []CommandDef{},
			wantLen:  64,
		},
		{
			name: "single command no options",
			commands: []CommandDef{
				{Name: "roll", Description: "Roll for a character"},
			},
			wantLen: 64,
		},
		{
			name: "single command with options",
			commands: []CommandDef{
				{
					Name:        "list",
					Description: "List characters",
					Options: []OptionDef{
						{Name: "user", Description: "User to list", Type: OptionUser, Required: false},
					},
				},
			},
			wantLen: 64,
		},
		{
			name: "command with choices",
			commands: []CommandDef{
				{
					Name:        "search",
					Description: "Search for something",
					Options: []OptionDef{
						{
							Name:        "type",
							Description: "Type of search",
							Type:        OptionString,
							Required:    true,
							Choices: []ChoiceDef{
								{Name: "anime", Value: "anime"},
								{Name: "manga", Value: "manga"},
							},
						},
					},
				},
			},
			wantLen: 64,
		},
		{
			name: "command with subcommands",
			commands: []CommandDef{
				{
					Name:        "profile",
					Description: "Profile commands",
					Options: []OptionDef{
						{
							Name:        "edit",
							Description: "Edit profile",
							Type:        OptionSubcommandGroup,
							Options: []OptionDef{
								{
									Name:        "quote",
									Description: "Set quote",
									Type:        OptionSubcommand,
									Options: []OptionDef{
										{Name: "value", Description: "Quote value", Type: OptionString, Required: true},
									},
								},
							},
						},
					},
				},
			},
			wantLen: 64,
		},
		{
			name: "int and float choices",
			commands: []CommandDef{
				{
					Name:        "test",
					Description: "Test",
					Options: []OptionDef{
						{
							Name:        "rating",
							Description: "Rating",
							Type:        OptionInt,
							Choices:     []ChoiceDef{{Name: "Low", Value: 1}, {Name: "High", Value: 5}},
						},
						{
							Name:        "score",
							Description: "Score",
							Type:        OptionNumber,
							Choices:     []ChoiceDef{{Name: "Half", Value: 0.5}},
						},
					},
				},
			},
			wantLen: 64,
		},
		{
			name: "all option types",
			commands: []CommandDef{
				{
					Name:        "alltypes",
					Description: "All types",
					Options: []OptionDef{
						{Name: "str", Description: "String", Type: OptionString},
						{Name: "int", Description: "Int", Type: OptionInt},
						{Name: "bool", Description: "Bool", Type: OptionBool},
						{Name: "user", Description: "User", Type: OptionUser},
						{Name: "channel", Description: "Channel", Type: OptionChannel},
						{Name: "role", Description: "Role", Type: OptionRole},
					},
				},
			},
			wantLen: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Hash(tt.commands)
			if len(got) != tt.wantLen {
				t.Errorf("Hash() length = %d, want %d", len(got), tt.wantLen)
			}
			for _, r := range got {
				if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
					t.Errorf("Hash() contains invalid character %q", r)
					break
				}
			}
		})
	}
}

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
