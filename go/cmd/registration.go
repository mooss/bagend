// This file implements subcommand registration.

package cmd

import "github.com/mooss/bagend/go/flag"

// Subcommand registers a subcommand.
// Registering subcommands to the same name overwrites the previous one.
func (c *Command) Subcommand(sub *Command) *Command {
	if _, exists := c.subs[sub.name]; !exists {
		c.suborder = append(c.suborder, sub.name)
	}

	c.subs[sub.name] = sub
	return sub
}

// Register creates and registers a subcommand, forwarding its arguments to Command.New.
func (c *Command) Register(name, docline string, opts ...flag.ParserOpt) *Command {
	return c.Subcommand(New(name, docline, opts...))
}
