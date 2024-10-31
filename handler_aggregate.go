package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/xristoskrik/gator/internal/database"
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

func handlerBrowse(s *State, cmd command, user database.User) error {
	limit := 2
	if len(os.Args) == 3 {
		if specifiedLimit, err := strconv.Atoi(os.Args[2]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}
	if len(os.Args) > 3 {
		return fmt.Errorf("browse <limit> (limit is optional)")
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}
