package main

import (
	"fmt"
	"os"
	"time"
)

func handlerAgg(s *State, cmd command) error {
	if len(os.Args) != 4 {
		return fmt.Errorf("usage: %v <time_between_reqs>", os.Stdin.Name())
	}

	timeBetweenRequests, err := time.ParseDuration(os.Args[3])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	fmt.Printf("Collecting feeds every %s...", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}
