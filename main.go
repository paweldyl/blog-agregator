package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"html"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	db   *database.Queries
	conf *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	availableCommands map[string]func(*state, command) error
	mu                *sync.Mutex
}

func main() {
	conf, err := config.Read()
	if err != nil {
		fmt.Println(err)
	}
	db, err := sql.Open("postgres", conf.ConnectString)
	if err != nil {
		fmt.Println("error while trying to connect to database")
	}
	dbQueries := database.New(db)

	newState := state{
		conf: &conf,
		db:   dbQueries,
	}

	newCommands := commands{
		availableCommands: map[string]func(*state, command) error{},
		mu:                &sync.Mutex{},
	}

	newCommands.register("login", handlerLogin)
	newCommands.register("register", handlerRegister)
	newCommands.register("reset", handlerReset)
	newCommands.register("users", handlerUsers)
	newCommands.register("agg", handlerAgg)
	newCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	newCommands.register("feeds", handlerFeeds)
	newCommands.register("follow", middlewareLoggedIn(handlerFollow))
	newCommands.register("following", middlewareLoggedIn(handlerFollowing))
	newCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	newCommands.register("browse", middlewareLoggedIn(handlerBrowse))

	passedArgs := os.Args

	if len(passedArgs) < 2 {
		fmt.Println("error: command required")
		os.Exit(1)
	}

	commandName := passedArgs[1]
	commandArgs := passedArgs[2:]
	calledCommand := command{
		name: commandName,
		args: commandArgs,
	}
	err = newCommands.run(&newState, calledCommand)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("expected username")
	}
	username := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return err
	}

	err = s.conf.SetUser(username)
	if err != nil {
		return err
	}

	fmt.Println("User has been set")

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("expected username")
	}
	username := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err == nil {
		return errors.New("user already exists")
	}

	//register user in db
	newUserParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	}
	dbUser, err := s.db.CreateUser(context.Background(), newUserParams)
	if err != nil {
		return err
	}

	err = s.conf.SetUser(username)
	if err != nil {
		return err
	}

	fmt.Println(dbUser)

	return nil
}

func handlerReset(s *state, _ command) error {
	_, err := s.db.CleanUsers(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func handlerUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	var printText string
	for _, user := range users {
		printText = "* " + user.Name
		if user.Name == s.conf.CurrentUserName {
			printText += " (current)"
		}
		fmt.Println(printText)
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("time_between_reqs required")
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Println("Collecting feeds every " + cmd.args[0])

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, dbUser database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("feed's name and url required")
	}
	feedName := cmd.args[0]
	feedUrl := cmd.args[1]
	ctx := context.Background()
	newFeedsParams := database.CreateFeedsParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    dbUser.ID,
	}
	dbFeed, err := s.db.CreateFeeds(ctx, newFeedsParams)
	if err != nil {
		return err
	}

	followCommandCmd := command{
		name: "follow",
		args: []string{
			feedUrl,
		},
	}
	err = handlerFollow(s, followCommandCmd, dbUser)
	if err != nil {
		return err
	}
	fmt.Println(dbFeed.ID)
	fmt.Println(dbFeed.CreatedAt)
	fmt.Println(dbFeed.UpdatedAt)
	fmt.Println(dbFeed.Name)
	fmt.Println(dbFeed.Url)
	fmt.Println(dbFeed.UserID)

	return nil
}

func handlerFeeds(s *state, _ command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(feed.UserName)
		fmt.Println("")
	}

	return nil
}

func handlerFollow(s *state, cmd command, dbUser database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("feed url required")
	}
	feedUrl := cmd.args[0]

	dbFeed, err := s.db.GetFeed(context.Background(), feedUrl)
	if err != nil {
		return err
	}
	newFeedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    dbUser.ID,
		FeedID:    dbFeed.ID,
	}
	feedFollowRes, err := s.db.CreateFeedFollow(context.Background(), newFeedFollow)
	if err != nil {
		return err
	}

	fmt.Print(feedFollowRes.UserName)
	fmt.Print(feedFollowRes.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command, dbUser database.User) error {
	feedFollowRes, err := s.db.GetFeedFollowsForUser(context.Background(), dbUser.Name)
	if err != nil {
		return err
	}
	for _, followedFeed := range feedFollowRes {
		fmt.Println(followedFeed.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, dbUser database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("url required")
	}
	feedUrl := cmd.args[0]
	deleteUserParams := database.DeleteFeedFollowsForUserParams{
		Name: dbUser.Name,
		Url:  feedUrl,
	}
	_, err := s.db.DeleteFeedFollowsForUser(context.Background(), deleteUserParams)
	if err != nil {
		return err
	}

	return nil
}

func handlerBrowse(s *state, cmd command, dbUser database.User) error {
	var limit int32 = 2
	if len(cmd.args) > 0 {
		limit = 2
	}
	getPostsParams := database.GetPostsForUserParams{
		Name:  dbUser.Name,
		Limit: limit,
	}
	posts, err := s.db.GetPostsForUser(context.Background(), getPostsParams)
	if err != nil {
		return err
	}
	for i, post := range posts {
		fmt.Println(post.Title)
		fmt.Println(post.Description)
		if i < len(posts)-1 {
			fmt.Println("-----------------------------------")
		}
	}

	return nil
}

func scrapeFeeds(s *state) error {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	nullNow := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	markFetchedParams := database.MarkFeedFetchedParams{
		LastFetchedAt: nullNow,
		UpdatedAt:     time.Now(),
		ID:            nextFeed.ID,
	}
	_, err = s.db.MarkFeedFetched(context.Background(), markFetchedParams)
	if err != nil {
		fmt.Println("err: " + err.Error())
		return err
	}
	rssFeed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return err
	}

	// fmt.Println(rssFeed.Channel.Title)
	// fmt.Println(rssFeed.Channel.Link)
	// fmt.Println(rssFeed.Channel.Description)
	for _, item := range rssFeed.Channel.Item {
		titleIsValid := item.Title != ""
		descIsValid := item.Description != ""
		nullTitle := sql.NullString{
			String: item.Title,
			Valid:  titleIsValid,
		}
		nullDesc := sql.NullString{
			String: item.Description,
			Valid:  descIsValid,
		}
		postData := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       nullTitle,
			Url:         item.Link,
			Description: nullDesc,
			PublishedAt: item.PubDate,
			FeedID:      nextFeed.ID,
		}
		_, err := s.db.CreatePost(context.Background(), postData)
		if err != nil && err.Error() != `pq: duplicate key value violates unique constraint "posts_url_key"` {
			fmt.Println(err)
		}
	}

	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		dbUser, err := s.db.GetUser(context.Background(), s.conf.CurrentUserName)
		if err != nil {
			return err
		}
		err = handler(s, cmd, dbUser)
		return err
	}
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.availableCommands[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.availableCommands[cmd.name](s, cmd)
	if err != nil {
		return err
	}
	return nil
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
	var returnRssFeed RSSFeed = RSSFeed{}
	request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "gator")

	client := http.DefaultClient

	httpRes, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer httpRes.Body.Close()

	byteRes, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(byteRes, &returnRssFeed)
	if err != nil {
		return nil, err
	}
	returnRssFeed.Channel.Title = html.UnescapeString(returnRssFeed.Channel.Title)
	returnRssFeed.Channel.Description = html.UnescapeString(returnRssFeed.Channel.Description)

	for i, item := range returnRssFeed.Channel.Item {
		returnRssFeed.Channel.Item[i].Title = html.UnescapeString(item.Title)
		returnRssFeed.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}

	return &returnRssFeed, nil
}
