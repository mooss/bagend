// Package cmd provides utilities to define commands and subcommands.
package cmd

import (
	"fmt"
	"strings"

	"github.com/mooss/bagend/go/flag"
	"github.com/mooss/bagend/go/fun/eager/lie"
	"github.com/mooss/bagend/go/obj/osl"
)

// This file defines the Command type, which groups flags and subcommands.

// Command groups a flag Parser with named subcommands.
// A Command can register flags through its Parser and other Commands.
type Command struct {
	// name is the command's name.
	Name string

	// Parent is the command that registered this command as a subcommand.
	Parent *Command

	// Docline is the command's one-line documentation.
	Docline string

	// Description is the command's longer description.
	Description string

	// Usage contains the command's usage arguments.
	Usage string

	// Flags stores the flags registered for the command.
	Flags *flag.Parser

	// Runner executes the command.
	Runner func() error

	// Alias contains alternate names for the command.
	Alias []string

	// subs maps each subcommand name to its Command.
	subs map[string]*Command

	// suborder records subcommand names in registration order.
	suborder []string
}

// New returns a Command with the given name and documentation line.
// opts are forwarded to the underlying flag.Parser.
func New(name, docline string) *Command {
	command := &Command{
		Name:    name,
		Docline: docline,
		subs:    map[string]*Command{},
	}
	command.Flags = flag.NewParser()
	return command
}

// AllNames returns a slice of the Command name and its aliases.
func (c *Command) AllNames() []string {
	return append([]string{c.Name}, c.Alias...)
}

// PipedNames returns a the Command name and its aliases joined by a pipe.
func (c *Command) PipedNames() string {
	return strings.Join(c.AllNames(), "|")
}

///////////////////////////
// Flags and subcommands //

// FlagsParser returns the underlying flag parser used to register flags.
func (c *Command) FlagsParser() *flag.Parser {
	return c.Flags
}

// Subcommands returns the registered subcommands in registration order.
func (c *Command) Subcommands() []*Command {
	return lie.Map(func(name string) *Command { return c.subs[name] }, c.suborder)
}

///////////////
// Help page //

// Help returns a formatted help page for the command and its direct subcommands.
func (c *Command) Help() string {
	c.ensureFlags()
	help := osl.New[string]()
	if c.Docline != "" {
		help.Push(c.Docline, "")
	}

	usage := c.commandPath()
	if c.Usage != "" {
		usage += " " + c.Usage
	}
	help.Push("Usage: " + usage)

	if c.Description != "" {
		help.Push("", c.Description)
	}

	flags := c.Flags.Help()
	if flags != "" {
		help.Push("", "Flags:", flags)
	}

	subcommands := c.subcommandsHelp()
	if subcommands != "" {
		help.Push("", "Commands:", subcommands)
	}
	help.Push("")

	return strings.Join(help, "\n")
}

func (c *Command) commandPath() string {
	if c.Parent == nil {
		return c.PipedNames()
	}
	return c.Parent.commandPath() + " " + c.PipedNames()
}

func (c *Command) subcommandsHelp() string {
	subcommands := c.Subcommands()
	if len(subcommands) == 0 {
		return ""
	}

	type helpline struct{ name, doc string }
	width := 0

	helps := lie.Map(func(sub *Command) helpline {
		name := sub.PipedNames()
		width = max(width, len(name))
		return helpline{name, sub.Docline}
	}, subcommands)

	lines := lie.Map(func(help helpline) string {
		return fmt.Sprintf("  %-*s  %s", width, help.name, help.doc)
	}, helps)

	return strings.Join(lines, "\n")
}

// ensureFlags ensures the flag parser is setup.
func (c *Command) ensureFlags() {
	if c.Flags == nil {
		c.Flags = flag.NewParser()
	}
	c.Flags.SetHelpPage(c.Help)
	c.Flags.EnableHelp()
}
