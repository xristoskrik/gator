package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/xristoskrik/gator/internal/database"
)

func handlerUsers(s *State, cmd command) error {
	fmt.Println("user command executing")
	usernames, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, item := range usernames {

		if item.Name == s.cfg.Username {
			fmt.Println(item.Name + " (current)")
			continue
		}
		fmt.Println(item.Name)

	}
	return nil
}
func handlerReset(s *State, cmd command) error {
	fmt.Println("reset command executing")

	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return err
	}

	return nil
}
func handlerLogin(s *State, cmd command) error {
	fmt.Println("login command executing")
	username := ""
	if len(cmd.args) != 3 {
		return fmt.Errorf("login <username>")
	}
	username = os.Args[2]
	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("user dont exist")
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Println("user has been set")
	return nil
}

func handlerRegister(s *State, cmd command) error {
	if len(cmd.args) != 3 {
		return fmt.Errorf("register <username>")
	}
	name := os.Args[2]

	test, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})

	if err != nil {
		return err
	}
	fmt.Println(test)
	err = s.cfg.SetUser(os.Args[2])
	if err != nil {
		return err
	}
	fmt.Println("user has been set")

	return nil
}
