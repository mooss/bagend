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

	// parser stores the flags registered for the command.
	parser *flag.Parser

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
func New(name, docline string, opts ...flag.ParserOpt) *Command {
	command := &Command{
		name:    name,
		docline: docline,
		parser:  flag.NewParser(opts...),
		subs:    map[string]*Command{},
	}
	command.parser.SetHelpPage(command.Help)
	return command
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) Docline() string {
	return c.docline
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
	c.parser.EnableHelp()
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

// Flags returns the underlying flag parser used to register flags.
func (c *Command) Flags() *flag.Parser {
	return c.parser
}

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

	usage := c.name
	if c.usage != "" {
		usage += " " + c.usage
	}
	help.Push("Usage: " + usage)

	if c.description != "" {
		help.Push("", c.description)
	}

	flags := c.parser.Help()
	if flags != "" {
		help.Push("", "Flags:", flags)
	}

	subcommands := c.subcommandsHelp()
	if subcommands != "" {
		help.Push("", "Commands:", subcommands)
	}

	return strings.Join(help, "\n")
}

func (c *Command) subcommandsHelp() string {
	subcommands := c.Subcommands()
	if len(subcommands) == 0 {
		return ""
	}

	width := 0
	for _, subcommand := range subcommands {
		width = max(width, len(subcommand.name))
	}

	lines := make([]string, len(subcommands))
	for i, subcommand := range subcommands {
		lines[i] = fmt.Sprintf("  %-*s  %s", width, subcommand.name, subcommand.docline)
	}

	return strings.Join(lines, "\n") + "\n"
}
