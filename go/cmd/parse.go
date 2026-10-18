// This file implements command dispatch to subcommands.

package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mooss/bagend/go/errs"
)

// Parse dispatches to the matching subcommand when the first argument names one,
// otherwise it parses the arguments with its own flag Parser.
func (c *Command) Parse(arguments []string) error {
	_, err := c.parse(arguments)
	return err
}

// Execute parses the arguments and runs the selected command.
func (c *Command) Execute(arguments []string) error {
	selected, err := c.parse(arguments)
	if err != nil {
		return err
	}
	if selected.runner == nil {
		return fmt.Errorf("command %q has no runner", selected.name)
	}
	return selected.runner()
}

func (c *Command) parse(arguments []string) (*Command, error) {
	if len(c.subs) > 0 && c.parser.AllowsPositionalArguments() {
		msg := "command %q cannot have both subcommands and positional arguments"
		return nil, fmt.Errorf(msg, c.name)
	}

	if sub, rest := c.findSub(arguments); sub != nil {
		return sub.parse(rest)
	}

	rest, stopped, err := c.parser.ParseUntil(arguments, func(arg string) bool {
		_, ok := c.lookup(arg)
		return ok
	})
	if err != nil {
		return nil, errs.Wrapf(err, "command %q", c.name)
	}
	if stopped {
		sub, _ := c.lookup(rest[0])
		return sub.parse(rest[1:])
	}

	return c, nil
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
	if sub, ok := c.subs[name]; ok {
		return sub, true
	}

	for _, cmd := range c.suborder {
		sub := c.subs[cmd]
		if slices.Contains(sub.aliases, name) {
			return sub, true
		}
	}

	return nil, false
}

func isFlag(arg string) bool {
	return strings.HasPrefix(arg, "-")
}
