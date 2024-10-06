package main

import (
	"errors"
	"fmt"
)

type command struct {
	command_name string
	args         []string
}

type Commands struct {
	cmd map[string]func(*State, command) error
}

func (c *Commands) register(name string, f func(*State, command) error) {
	c.cmd[name] = f
}

func (c *Commands) run(s *State, cmd command) error {
	fmt.Printf("Executing command: %s\n", cmd.command_name)
	if handler, exists := c.cmd[cmd.command_name]; exists {
		return handler(s, cmd)
	}
	return errors.New("handler dont exist")
}
