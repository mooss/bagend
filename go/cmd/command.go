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
	name string

	// docline is the command's one-line documentation.
	docline string

	// description is the command's longer description.
	description string

	// usage contains the command's usage arguments.
	usage string

	// Flags stores the flags registered for the command.
	Flags *flag.Parser

	// runner executes the command.
	runner func() error

	// subs maps each subcommand name to its Command.
	subs map[string]*Command

	// aliases contains alternate names for the command.
	aliases []string

	// suborder records subcommand names in registration order.
	suborder []string
}

// New returns a Command with the given name and documentation line.
// opts are forwarded to the underlying flag.Parser.
func New(name, docline string) *Command {
	command := &Command{
		name:    name,
		docline: docline,
		Flags:   flag.NewParser(),
		subs:    map[string]*Command{},
	}
	command.Flags.SetHelpPage(command.Help)
	command.Flags.EnableHelp()
	return command
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) Docline() string {
	return c.docline
}

// AllNames returns a slice of the Command name and its aliases.
func (c *Command) AllNames() []string {
	return append([]string{c.name}, c.aliases...)
}

// PipedNames returns a the Command name and its aliases joined by a pipe.
func (c *Command) PipedNames() string {
	return strings.Join(c.AllNames(), "|")
}

//////////////////////
// Fluent interface //

// Description sets the command's longer description and returns it for fluent configuration.
func (c *Command) Description(description string) *Command {
	c.description = description
	return c
}

// Usage sets the command's usage arguments and returns it for fluent configuration.
func (c *Command) Usage(usage string) *Command {
	c.usage = usage
	return c
}

// Alias adds aliases for this command and returns it for fluent configuration.
func (c *Command) Alias(aliases ...string) *Command {
	c.aliases = append(c.aliases, aliases...)
	return c
}

// Runner sets the function called to run the command.
func (c *Command) Runner(runner func() error) *Command {
	c.runner = runner
	return c
}

///////////////////////////
// Flags and subcommands //

// Subcommands returns the registered subcommands in registration order.
func (c *Command) Subcommands() []*Command {
	return lie.Map(func(name string) *Command { return c.subs[name] }, c.suborder)
}

///////////////
// Help page //

// Help returns a formatted help page for the command and its direct subcommands.
func (c *Command) Help() string {
	help := osl.New[string]()
	if c.docline != "" {
		help.Push(c.docline, "")
	}

	usage := c.PipedNames()
	if c.usage != "" {
		usage += " " + c.usage
	}
	help.Push("Usage: " + usage)

	if c.description != "" {
		help.Push("", c.description)
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
		return helpline{name, sub.docline}
	}, subcommands)

	lines := lie.Map(func(help helpline) string {
		return fmt.Sprintf("  %-*s  %s", width, help.name, help.doc)
	}, helps)

	return strings.Join(lines, "\n")
}
