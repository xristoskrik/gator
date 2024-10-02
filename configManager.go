package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/xristoskrik/gator/internal/database"
)

const configFileName = ".gatorconfig.json"

type State struct {
	db  *database.Queries
	cfg *Config
}
type command struct {
	command_name string
	args         []string
}

func handlerLogin(s *State, cmd command) error {
	fmt.Println("login command executing")
	username := ""
	if len(cmd.args) != 3 {
		return fmt.Errorf("login <username>")
	}
	username = os.Args[2]
	exists, err := s.db.UserExists(context.Background(), username)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("user dont exist")
	}
	err = s.cfg.SetUser(os.Args[2])
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

type Config struct {
	Db_url   string `db_url:"url"`
	Username string `current_user_name:"username"`
}

func (c *Config) PrintData() {
	fmt.Println(c.Db_url)
	fmt.Println(c.Username)
}
func getConfigFilePath() (string, error) {

	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homedir + "/" + configFileName, nil

}

func (c *Config) SetUser(username string) error {
	c.Username = username

	jsonDat, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	err = os.WriteFile(path, jsonDat, 0644)
	if err != nil {
		return err
	}
	return nil
}
func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, err
}
