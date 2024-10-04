package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
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
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}

	req.Header.Add("User-Agent", "gator")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	// Unescape HTML entities in each item's Title and Description
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}
	return &feed, nil

}
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
func handlerAgg(s *State, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	for i := range feed.Channel.Item {
		fmt.Println(feed.Channel.Item[i].Title)
		fmt.Println(feed.Channel.Item[i].Description)
	}
	return nil
}
func createFeedFollow(s *State, user_id uuid.UUID, feed_id uuid.UUID) (database.CreateFeedFollowRow, error) {
	feed_follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user_id,
		FeedID:    feed_id,
	})
	if err != nil {
		return database.CreateFeedFollowRow{}, err
	}

	return feed_follow, nil
}
func handlerFollow(s *State, cmd command, user database.User) error {
	if len(os.Args) != 3 {
		return fmt.Errorf("follow <url>")
	}

	url := os.Args[2]
	feed, err := s.db.GetFeedFromUrl(context.Background(), url)
	if err != nil {
		return err
	}
	feed_follow, err := createFeedFollow(s, user.ID, feed.ID)
	if err != nil {
		return err
	}
	fmt.Println(feed_follow)
	return nil
}
func handlerFeeds(s *State, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, item := range feeds {
		fmt.Printf("%s\n%s\n%s\n", item.Name, item.Url, item.Username.String)
	}
	return nil
}
func handlerUnfollow(s *State, cmd command, user database.User) error {
	if len(cmd.args) != 3 {
		return fmt.Errorf("unfollow <url>")
	}
	url := os.Args[2]
	err := s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		Url:    url,
	})
	if err != nil {
		return err
	}
	return nil
}
func handlerAddFeed(s *State, cmd command, user database.User) error {
	fmt.Println("add feed command executing")
	if len(cmd.args) != 4 {
		return fmt.Errorf("addfeed <name><url>")
	}

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      os.Args[2],
		Url:       os.Args[3],
		UserID:    user.ID,
	})
	if err != nil {
		return err
	}
	fmt.Println(feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.Url)
	_, err = createFeedFollow(s, user.ID, feed.ID)
	if err != nil {
		return err
	}
	return nil
}
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
