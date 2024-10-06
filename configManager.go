package main

import (
	"encoding/json"

	"fmt"

	"io"

	"os"

	_ "github.com/lib/pq"
)

const configFileName = ".gatorconfig.json"

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
