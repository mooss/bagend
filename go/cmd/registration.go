// This file implements subcommand registration.

package cmd

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
func (c *Command) Register(name, docline string) *Command {
	return c.Subcommand(New(name, docline))
}

// RegisterAll registers all the given commands as subcommands.
func (c *Command) RegisterAll(subs ...*Command) {
	for _, sub := range subs {
		c.Subcommand(sub)
	}
}
