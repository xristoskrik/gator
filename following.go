package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xristoskrik/gator/internal/database"
)

func handlerFollowing(s *State, cmd command, user database.User) error {
	if len(os.Args) != 2 {
		return fmt.Errorf("you dont need parameters")
	}

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	for _, item := range feeds {
		fmt.Println(item.FeedName)
	}

	return nil
}
