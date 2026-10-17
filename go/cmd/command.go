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
	name    string
	docline string
	parser  *flag.Parser
	runner  func() error
	subs    map[string]*Command
	order   []string
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
	return lie.Map(func(name string) *Command { return c.subs[name] }, c.order)
}
