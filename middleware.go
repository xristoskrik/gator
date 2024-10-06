package main

import (
	"context"
	"fmt"

	"github.com/xristoskrik/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *State, cmd command, user database.User) error) func(*State, command) error {
	return func(s *State, cmd command) error {
		// Retrieve the user from the session or request context
		user, err := s.db.GetUser(context.Background(), s.cfg.Username)
		if err != nil {
			// If there's no logged-in user, return an appropriate error
			return fmt.Errorf("user must be logged in to access this resource")
		}

		// Call the original handler with the user
		return handler(s, cmd, user)
	}
}
