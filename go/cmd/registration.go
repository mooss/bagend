// This file implements subcommand registration.

package cmd

// RegisterSub registers a subcommand.
// Registering subcommands to the same name overwrites the previous one.
func (c *Command) RegisterSub(sub *Command) *Command {
	if _, exists := c.subs[sub.name]; !exists {
		c.suborder = append(c.suborder, sub.name)
	}

	c.subs[sub.name] = sub
	return sub
}

// RegisterNewSub creates and registers a subcommand, forwarding its arguments to Command.New.
func (c *Command) RegisterNewSub(name, docline string) *Command {
	return c.RegisterSub(New(name, docline))
}

// RegisterSubs registers all the given commands as subcommands.
func (c *Command) RegisterSubs(subs ...*Command) {
	for _, sub := range subs {
		c.RegisterSub(sub)
	}
}
