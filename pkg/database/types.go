package database

import (
	"database/sql"
	"time"
)

// Config holds all the configuration for the db
type Config struct {
	host     string
	port     int
	user     string
	password string
	dbname   string
	sslmode  string
}

// Database contains all the methods to interact with DB
type Database struct {
	config *Config
	db     *sql.DB
}

// Color is a color legend
type Color struct {
	Color  string `json:"color"`
	Legend string `json:"legend"`
}

// StoryboardUser aka user
type StoryboardUser struct {
	UserID   string `json:"id"`
	UserName string `json:"name"`
	Active   bool   `json:"active"`
}

// Storyboard A story mapping board
type Storyboard struct {
	StoryboardID   string               `json:"id"`
	OwnerID        string               `json:"owner_id"`
	StoryboardName string               `json:"name"`
	Users          []*StoryboardUser    `json:"users"`
	Goals          []*StoryboardGoal    `json:"goals"`
	ColorLegend    []*Color             `json:"color_legend"`
	Personas       []*StoryboardPersona `json:"personas"`
}

// StoryboardGoal A row in a story mapping board
type StoryboardGoal struct {
	GoalID    string              `json:"id"`
	GoalName  string              `json:"name"`
	Columns   []*StoryboardColumn `json:"columns"`
	SortOrder int                 `json:"sort_order"`
}

// StoryboardColumn A column in a storyboard goal
type StoryboardColumn struct {
	ColumnID   string             `json:"id"`
	ColumnName string             `json:"name"`
	Stories    []*StoryboardStory `json:"stories"`
	SortOrder  int                `json:"sort_order"`
}

// StoryboardStory A story in a storyboard goal column
type StoryboardStory struct {
	StoryID      string          `json:"id"`
	StoryName    string          `json:"name"`
	StoryContent string          `json:"content"`
	StoryColor   string          `json:"color"`
	StoryPoints  int             `json:"points"`
	StoryClosed  bool            `json:"closed"`
	SortOrder    int             `json:"sort_order"`
	Comments     []*StoryComment `json:"comments"`
}

// StoryComment A story comment by a user
type StoryComment struct {
	StoryID    string `json:"story_id"`
	UserID     string `json:"user_id"`
	Comment    string `json:"comment"`
	CreateDate string `json:"created_date"`
}

// StoryboardPersona A storyboards personas
type StoryboardPersona struct {
	PersonaID   string `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Description string `json:"description"`
}

// User aka user
type User struct {
	UserID     string `json:"id"`
	UserName   string `json:"name"`
	UserEmail  string `json:"email"`
	UserAvatar string `json:"avatar"`
	UserType   string `json:"type"`
	Verified   bool   `json:"verified"`
}

// APIKey structure
type APIKey struct {
	ID          string    `json:"id"`
	Prefix      string    `json:"prefix"`
	UserID      string    `json:"userId"`
	Name        string    `json:"name"`
	Key         string    `json:"apiKey"`
	Active      bool      `json:"active"`
	CreatedDate time.Time `json:"createdDate"`
	UpdatedDate time.Time `json:"updatedDate"`
}
