// This file implements subcommand registration.

package cmd

// Add registers a subcommand.
// Registering subcommands to the same name overwrites the previous one.
func (c *Command) Add(sub *Command) {
	if _, exists := c.subs[sub.name]; !exists {
		c.suborder = append(c.suborder, sub.name)
	}

	c.subs[sub.name] = sub
}
