package uconfig

import "fmt"

func (c *config) FormattedUsage(format string) {
	switch format {
	case "env":
		c.usageEnv()
	case "yaml":
		c.usageYaml()
	default:
		panic("invalid format for printing config usage: " + format)
	}
}

func (c *config) usageEnv() {
	for _, f := range c.fields {
		fmt.Fprintf(UsageOutput, "%s=%s\n", f.Meta()["env"], f.Meta()["default"])
	}
}

func (c *config) usageYaml() {
	for _, f := range c.fields {
		fmt.Fprintf(UsageOutput, "%s: \"%s\"\n", f.Meta()["env"], f.Meta()["default"])
	}
}
