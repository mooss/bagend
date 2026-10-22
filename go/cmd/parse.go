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
	_, err := c.parse(arguments, nil)
	return err
}

// Execute parses the arguments and runs the selected command.
func (c *Command) Execute(arguments []string) error {
	selected, err := c.parse(arguments, nil)
	if err != nil {
		return err
	}
	if selected.Runner == nil {
		return fmt.Errorf("command %q has no runner", selected.Name)
	}
	return selected.Runner()
}

func (c *Command) parse(arguments []string, path []string) (*Command, error) {
	c.ensureFlags()

	path = append(path, c.Name)
	if err := c.validateSubcommands(path); err != nil {
		return nil, err
	}

	if sub, rest := c.findSub(arguments); sub != nil {
		return sub.parse(rest, path)
	}

	return c.parseArguments(arguments, path)
}

func (c *Command) validateSubcommands(path []string) error {
	if len(c.subs) > 0 && c.Flags.AllowsPositionalArguments() {
		return fmt.Errorf(
			"command %s cannot have both subcommands and positional arguments",
			strings.Join(path, " "),
		)
	}
	return nil
}

// Invariant: c.Flags must be initialized.
func (c *Command) parseArguments(arguments []string, path []string) (*Command, error) {
	rest, stopped, err := c.Flags.ParseUntil(arguments, c.shouldStop)
	if err != nil {
		return nil, errs.Wrapf(err, "command %s", strings.Join(path, " "))
	}

	if stopped {
		sub, ok := c.lookup(rest[0])
		if !ok {
			err := fmt.Errorf("invalid subcommand %q", rest[0])
			return nil, errs.Wrapf(err, "command %s", strings.Join(path, " "))
		}
		return sub.parse(rest[1:], path)
	}

	return c, nil
}

func (c *Command) shouldStop(arg string) bool {
	if len(c.subs) > 0 {
		return true
	}
	_, ok := c.lookup(arg)
	return ok
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
		if sub != nil && slices.Contains(sub.Aliases, name) {
			return sub, true
		}
	}

	return nil, false
}

func isFlag(arg string) bool {
	return strings.HasPrefix(arg, "-")
}
