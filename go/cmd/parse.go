// This file implements command dispatch to subcommands.

package cmd

import "strings"

// Parse dispatches to the matching subcommand when the first argument names one,
// otherwise it parses the arguments with its own flag Parser.
func (c *Command) Parse(arguments []string) error {
	if sub, rest := c.findSub(arguments); sub != nil {
		return sub.Parse(rest)
	}

	return c.parser.Parse(arguments)
}

// findSub returns the subcommand named by the first argument and the remaining arguments.
func (c *Command) findSub(arguments []string) (*Command, []string) {
	if len(arguments) == 0 {
		return nil, nil
	}

	if isFlag(arguments[0]) {
		return nil, nil
	}

	sub, ok := c.lookup(arguments[0])
	if !ok {
		return nil, nil
	}

	return sub, arguments[1:]
}

func (c *Command) lookup(name string) (*Command, bool) {
	sub, ok := c.subs[name]
	return sub, ok
}

func isFlag(arg string) bool {
	return strings.HasPrefix(arg, "-")
}
