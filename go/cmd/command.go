// Package cmd provides utilities to define commands and subcommands.
package cmd

import (
	"github.com/mooss/bagend/go/flag"
	"github.com/mooss/bagend/go/fun/eager/lie"
)

// This file defines the Command type, which groups flags and subcommands.

// Command groups a flag Parser with named subcommands.
// A Command can register flags through its Parser and other Commands.
type Command struct {
	// name is the command's name.
	name string

	// docline is the command's one-line documentation.
	docline string

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
	return &Command{
		name:    name,
		docline: docline,
		parser:  flag.NewParser(opts...),
		subs:    map[string]*Command{},
	}
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) Docline() string {
	return c.docline
}

// Alias adds aliases for this command and returns it for fluent configuration.
func (c *Command) Alias(aliases ...string) *Command {
	c.aliases = append(c.aliases, aliases...)
	return c
}

// Flags returns the underlying flag parser used to register flags.
func (c *Command) Flags() *flag.Parser {
	return c.parser
}

// Runner sets the function called to run the command.
func (c *Command) Runner(runner func() error) *Command {
	c.runner = runner
	return c
}

// Subcommands returns the registered subcommands in registration order.
func (c *Command) Subcommands() []*Command {
	return lie.Map(func(name string) *Command { return c.subs[name] }, c.suborder)
}
