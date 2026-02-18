package discord

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"

	"cmp"

	"github.com/Karitham/corde"
)

type CommandDef struct {
	Name        string
	Description string
	Options     []OptionDef
}

type OptionDef struct {
	Name         string
	Description  string
	Type         OptionType
	Required     bool
	Autocomplete bool
	Choices      []ChoiceDef
	Options      []OptionDef
}

type ChoiceDef struct {
	Name  string
	Value any // string, int, or float64
}

type OptionType int

const (
	OptionSubcommandGroup OptionType = iota + 1
	OptionSubcommand
	OptionString
	OptionInt
	OptionBool
	OptionUser
	OptionChannel
	OptionRole
	OptionMentionable
	OptionNumber
)

func Hash(cmds []CommandDef) string {
	sorted := slices.Clone(cmds)
	sortCommands(sorted)

	b, _ := json.Marshal(sorted)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func sortCommands(cmds []CommandDef) {
	slices.SortFunc(cmds, func(a, b CommandDef) int {
		return cmp.Compare(a.Name, b.Name)
	})
	for i := range cmds {
		sortOptions(cmds[i].Options)
	}
}

func sortOptions(opts []OptionDef) {
	slices.SortFunc(opts, func(a, b OptionDef) int {
		return cmp.Compare(a.Name, b.Name)
	})
	for i := range opts {
		slices.SortFunc(opts[i].Choices, func(a, b ChoiceDef) int {
			return cmp.Compare(a.Name, b.Name)
		})
		sortOptions(opts[i].Options)
	}
}

func (c CommandDef) ToCorde() corde.CreateCommander {
	opts := make([]corde.CreateOptioner, len(c.Options))
	for i, o := range c.Options {
		opts[i] = o.ToCorde()
	}
	return corde.NewSlashCommand(c.Name, c.Description, opts...)
}

func (o OptionDef) ToCorde() corde.CreateOptioner {
	switch o.Type {
	case OptionSubcommandGroup:
		opts := make([]corde.CreateOptioner, len(o.Options))
		for i, opt := range o.Options {
			opts[i] = opt.ToCorde()
		}
		return corde.NewSubcommandGroup(o.Name, o.Description, opts...)
	case OptionSubcommand:
		opts := make([]corde.CreateOptioner, len(o.Options))
		for i, opt := range o.Options {
			opts[i] = opt.ToCorde()
		}
		return corde.NewSubcommand(o.Name, o.Description, opts...)
	case OptionString:
		choices := make([]corde.Choice[string], len(o.Choices))
		for i, c := range o.Choices {
			choices[i] = corde.Choice[string]{Name: c.Name, Value: c.Value.(string)}
		}
		opt := corde.NewStringOption(o.Name, o.Description, o.Required, choices...)
		if o.Autocomplete {
			opt.CanAutocomplete()
		}
		return opt
	case OptionInt:
		choices := make([]corde.Choice[int], len(o.Choices))
		for i, c := range o.Choices {
			choices[i] = corde.Choice[int]{Name: c.Name, Value: int(c.Value.(int))}
		}
		opt := corde.NewIntOption(o.Name, o.Description, o.Required, choices...)
		if o.Autocomplete {
			opt.CanAutocomplete()
		}
		return opt
	case OptionBool:
		return corde.NewBoolOption(o.Name, o.Description, o.Required)
	case OptionUser:
		return corde.NewUserOption(o.Name, o.Description, o.Required)
	case OptionChannel:
		return corde.NewChannelOption(o.Name, o.Description, o.Required)
	case OptionRole:
		return corde.NewRoleOption(o.Name, o.Description, o.Required)
	case OptionMentionable:
		return corde.NewMentionableOption(o.Name, o.Description, o.Required)
	case OptionNumber:
		choices := make([]corde.Choice[float64], len(o.Choices))
		for i, c := range o.Choices {
			choices[i] = corde.Choice[float64]{Name: c.Name, Value: c.Value.(float64)}
		}
		opt := corde.NewNumberOption(o.Name, o.Description, o.Required, choices...)
		if o.Autocomplete {
			opt.CanAutocomplete()
		}
		return opt
	default:
		return corde.NewStringOption(o.Name, o.Description, o.Required)
	}
}

func ToCorde(cmds []CommandDef) []corde.CreateCommander {
	result := make([]corde.CreateCommander, len(cmds))
	for i, c := range cmds {
		result[i] = c.ToCorde()
	}
	return result
}
