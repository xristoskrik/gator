package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/xristoskrik/gator/internal/database"
)

func main() {

	config, err := Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	state := State{
		cfg: &config,
	}
	commands := Commands{cmd: make(map[string]func(*State, command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddFeed)
	commands.register("feeds", handlerFeeds)
	db, err := sql.Open("postgres", state.cfg.Db_url)
	state.db = database.New(db)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}
	if len(os.Args) == 1 {
		fmt.Println("err: no arguments passed")
		os.Exit(1)
	}
	err = commands.run(&state, command{command_name: os.Args[1], args: os.Args})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
