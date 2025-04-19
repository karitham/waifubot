package main

import (
	"flag"
	"fmt"
	"strings"
	"syscall"

	"github.com/Karitham/corde"
)

type cliSnowflake struct {
	Dest       *corde.Snowflake
	EnvVars    []string
	Required   bool
	HasBeenSet bool
}

func (c *cliSnowflake) Snowflake() *corde.Snowflake {
	return c.Dest
}

func (c *cliSnowflake) String() string {
	if len(c.EnvVars) < 1 {
		return ""
	}
	envStr := strings.Builder{}
	envStr.WriteRune('[')
	for _, env := range c.EnvVars {
		if envStr.Len() > 1 {
			envStr.WriteString(", ")
		}
		envStr.WriteRune('$')
		envStr.WriteString(env)
	}
	envStr.WriteRune(']')

	return fmt.Sprintf("--%s\t(default: %s) %s", c.EnvVars[0], c.Dest.String(), envStr.String())
}

// Apply Flag settings to the given flag set
func (c *cliSnowflake) Apply(f *flag.FlagSet) error {
	for _, envVar := range c.EnvVars {
		if v, ok := syscall.Getenv(envVar); ok {
			snow := corde.SnowflakeFromString(v)
			if snow == 0 {
				return fmt.Errorf("invalid snowflake %s", v)
			}

			c.Dest = &snow
			c.HasBeenSet = true
		}
	}
	if c.Required && c.Dest == nil {
		return fmt.Errorf("required flag %s not set", c.EnvVars[0])
	}
	return nil
}

func (c *cliSnowflake) Names() []string {
	return c.EnvVars
}

func (c *cliSnowflake) IsSet() bool {
	return c.HasBeenSet
}
