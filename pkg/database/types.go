package database

import "database/sql"

// Config holds all the configuration for the db
type Config struct {
	host     string
	port     int
	user     string
	password string
	dbname   string
}

// Database contains all the methods to interact with DB
type Database struct {
	config *Config
	db     *sql.DB
}

// StoryboardUser aka user
type StoryboardUser struct {
	UserID   string `json:"id"`
	UserName string `json:"name"`
	Active   bool   `json:"active"`
}

// Storyboard A story mapping board
type Storyboard struct {
	StoryboardID   string            `json:"id"`
	OwnerID        string            `json:"ownerId"`
	StoryboardName string            `json:"name"`
	Users          []*StoryboardUser `json:"users"`
	Goals          []*StoryboardGoal `json:"goals"`
}

// StoryboardGoal A row in a story mapping board
type StoryboardGoal struct {
	GoalID    string              `json:"id"`
	GoalName  string              `json:"name"`
	Columns   []*StoryboardColumn `json:"columns"`
	SortOrder int                 `json:"sortOrder"`
}

// StoryboardColumn A column in a storyboard goal
type StoryboardColumn struct {
	ColumnID   string             `json:"id"`
	ColumnName string             `json:"name"`
	Stories    []*StoryboardStory `json:"stories"`
	SortOrder  int                `json:"sortOrder"`
}

// StoryboardStory A story in a storyboard goal column
type StoryboardStory struct {
	StoryID      string `json:"id"`
	StoryName    string `json:"name"`
	StoryContent string `json:"content"`
	StoryColor   string `json:"color"`
	SortOrder    int    `json:"sortOrder"`
}

// User aka user
type User struct {
	UserID    string `json:"id"`
	UserName  string `json:"name"`
	UserEmail string `json:"email"`
	UserType  string `json:"type"`
	Verified  bool   `json:"verified"`
}
