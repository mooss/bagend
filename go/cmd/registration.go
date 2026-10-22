// This file implements subcommand registration.

package cmd

// RegisterSub registers a subcommand.
// Registering subcommands to the same name overwrites the previous one.
func (c *Command) RegisterSub(sub *Command) *Command {
	if c.subs == nil {
		c.subs = map[string]*Command{}
	}

	if _, exists := c.subs[sub.Name]; !exists {
		c.suborder = append(c.suborder, sub.Name)
	}

	sub.Parent = c
	c.subs[sub.Name] = sub
	return sub
}

// RegisterSubs registers all the given commands as subcommands.
func (c *Command) RegisterSubs(subs ...*Command) {
	for _, sub := range subs {
		c.RegisterSub(sub)
	}
}
