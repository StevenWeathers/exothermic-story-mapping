package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/lib/pq"
	"github.com/markbates/pkger"
)

var db *sql.DB

// Storyboard A story mapping board
type Storyboard struct {
	StoryboardID   string            `json:"id"`
	OwnerID        string            `json:"ownerId"`
	StoryboardName string            `json:"name"`
	Users          []*User           `json:"users"`
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
	Active    bool   `json:"active"` // this is actually for storyboard active status
}

// SetupDB runs db migrations, sets up a db connection pool
// and sets previously active users to false during startup
func SetupDB() {
	var (
		host     = GetEnv("DB_HOST", "db")
		port     = GetIntEnv("DB_PORT", 5432)
		user     = GetEnv("DB_USER", "thor")
		password = GetEnv("DB_PASS", "odinson")
		dbname   = GetEnv("DB_NAME", "exothermic")
	)

	sqlFile, ioErr := pkger.Open("/schema.sql")
	if ioErr != nil {
		log.Println("Error reading schema.sql file required to migrate db")
		log.Fatal(ioErr)
	}
	sqlContent, ioErr := ioutil.ReadAll(sqlFile)
	if ioErr != nil {
		// this will hopefully only possibly panic during development as the file is already in memory otherwise
		log.Println("Error reading schema.sql file required to migrate db")
		log.Fatal(ioErr)
	}
	migrationSQL := string(sqlContent)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	if _, err := db.Exec(migrationSQL); err != nil {
		log.Fatal(err)
	}

	// on server start reset all users to active false for storyboards
	if _, err := db.Exec(
		`call deactivate_all_users();`); err != nil {
		log.Println(err)
	}

	// on server start if admin email is specified set that user to ADMIN type
	if AdminEmail != "" {
		if _, err := db.Exec(
			`call promote_user_by_email($1);`,
			AdminEmail,
		); err != nil {
			log.Println(err)
		}
	}
}

//CreateStoryboard adds a new storyboard to the db
func CreateStoryboard(OwnerID string, StoryboardName string) (*Storyboard, error) {
	var b = &Storyboard{
		StoryboardID:   "",
		OwnerID:        OwnerID,
		StoryboardName: StoryboardName,
		Users:          make([]*User, 0),
	}

	e := db.QueryRow(
		`SELECT * FROM create_storyboard($1, $2);`,
		OwnerID,
		StoryboardName,
	).Scan(&b.StoryboardID)
	if e != nil {
		log.Println(e)
		return nil, errors.New("Error Creating Storyboard")
	}

	return b, nil
}

// GetStoryboard gets a storyboard by ID
func GetStoryboard(StoryboardID string) (*Storyboard, error) {
	var b = &Storyboard{
		StoryboardID:   StoryboardID,
		OwnerID:        "",
		StoryboardName: "",
		Users:          make([]*User, 0),
		Goals:          make([]*StoryboardGoal, 0),
	}

	// get storyboard
	e := db.QueryRow(
		"SELECT id, name, owner_id FROM storyboard WHERE id = $1",
		StoryboardID,
	).Scan(
		&b.StoryboardID,
		&b.StoryboardName,
		&b.OwnerID,
	)
	if e != nil {
		log.Println(e)
		return nil, errors.New("Not found")
	}

	b.Users = GetStoryboardUsers(StoryboardID)
	b.Goals = GetStoryboardGoals(StoryboardID)

	return b, nil
}

// GetStoryboardsByUser gets a list of storyboards by UserID
func GetStoryboardsByUser(UserID string) ([]*Storyboard, error) {
	var storyboards = make([]*Storyboard, 0)
	storyboardRows, storyboardsErr := db.Query(`
		SELECT * FROM get_storyboards_by_user($1);
	`, UserID)
	if storyboardsErr != nil {
		return nil, errors.New("Not found")
	}

	defer storyboardRows.Close()
	for storyboardRows.Next() {
		var b = &Storyboard{
			StoryboardID:   "",
			OwnerID:        "",
			StoryboardName: "",
			Users:          make([]*User, 0),
		}
		if err := storyboardRows.Scan(
			&b.StoryboardID,
			&b.StoryboardName,
			&b.OwnerID,
		); err != nil {
			log.Println(err)
		} else {
			storyboards = append(storyboards, b)
		}
	}

	return storyboards, nil
}

// ConfirmOwner confirms the user is infact owner of the storyboard
func ConfirmOwner(StoryboardID string, userID string) error {
	var ownerID string
	e := db.QueryRow("SELECT owner_id FROM storyboard WHERE id = $1", StoryboardID).Scan(&ownerID)
	if e != nil {
		log.Println(e)
		return errors.New("Storyboard Not found")
	}

	if ownerID != userID {
		return errors.New("Not Owner")
	}

	return nil
}

// GetStoryboardUser gets a user from db by ID and checks storyboard active status
func GetStoryboardUser(StoryboardID string, UserID string) (*User, error) {
	var active bool
	var w User

	e := db.QueryRow(
		`SELECT * FROM get_storyboard_user($1, $2);`,
		StoryboardID,
		UserID,
	).Scan(
		&w.UserID,
		&w.UserName,
		&w.UserEmail,
		&w.UserType,
		&active,
	)
	if e != nil {
		log.Println(e)
		return nil, errors.New("User Not found")
	}

	if active {
		return nil, errors.New("User Already Active in Storyboard")
	}

	return &w, nil
}

// GetStoryboardUsers retrieves the users for a given storyboard from db
func GetStoryboardUsers(StoryboardID string) []*User {
	var users = make([]*User, 0)
	rows, err := db.Query(
		`SELECT * FROM get_storyboard_users($1);`,
		StoryboardID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var w User
			if err := rows.Scan(&w.UserID, &w.UserName, &w.UserEmail, &w.UserType, &w.Active); err != nil {
				log.Println(err)
			} else {
				users = append(users, &w)
			}
		}
	}

	return users
}

// AddUserToStoryboard adds a user by ID to the storyboard by ID
func AddUserToStoryboard(StoryboardID string, UserID string) ([]*User, error) {
	if _, err := db.Exec(
		`INSERT INTO storyboard_user (storyboard_id, user_id, active)
		VALUES ($1, $2, true)
		ON CONFLICT (storyboard_id, user_id) DO UPDATE SET active = true`,
		StoryboardID,
		UserID,
	); err != nil {
		log.Println(err)
	}

	users := GetStoryboardUsers(StoryboardID)

	return users, nil
}

// RetreatUser removes a user from the current storyboard by ID
func RetreatUser(StoryboardID string, UserID string) []*User {
	if _, err := db.Exec(
		`UPDATE storyboard_user SET active = false WHERE storyboard_id = $1 AND user_id = $2`, StoryboardID, UserID); err != nil {
		log.Println(err)
	}

	if _, err := db.Exec(
		`UPDATE users SET last_active = NOW() WHERE id = $1`, UserID); err != nil {
		log.Println(err)
	}

	users := GetStoryboardUsers(StoryboardID)

	return users
}

// SetStoryboardOwner sets the ownerId for the storyboard
func SetStoryboardOwner(StoryboardID string, userID string, OwnerID string) (*Storyboard, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call set_storyboard_owner($1, $2);`, StoryboardID, OwnerID); err != nil {
		log.Println(err)
	}

	storyboard, err := GetStoryboard(StoryboardID)
	if err != nil {
		return nil, errors.New("Unable to promote owner")
	}

	return storyboard, nil
}

// DeleteStoryboard removes all storyboard associations and the storyboard itself from DB by StoryboardID
func DeleteStoryboard(StoryboardID string, userID string) error {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call delete_storyboard($1);`, StoryboardID); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

/*
	Storyboard Goal
*/

// CreateStoryboardGoal adds a new goal to a Storyboard
func CreateStoryboardGoal(StoryboardID string, userID string, GoalName string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call create_storyboard_goal($1, $2);`, StoryboardID, GoalName,
	); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// ReviseGoalName updates the plan name by ID
func ReviseGoalName(StoryboardID string, userID string, GoalID string, GoalName string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call update_storyboard_goal($1, $2);`,
		GoalID,
		GoalName,
	); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// DeleteStoryboardGoal removes a goal from the current board by ID
func DeleteStoryboardGoal(StoryboardID string, userID string, GoalID string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call delete_storyboard_goal($1);`, GoalID); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// GetStoryboardGoals retrieves goals for given storyboard from db
func GetStoryboardGoals(StoryboardID string) []*StoryboardGoal {
	var goals = make([]*StoryboardGoal, 0)

	goalRows, goalsErr := db.Query(
		`SELECT * FROM get_storyboard_goals($1);`,
		StoryboardID,
	)
	if goalsErr == nil {
		defer goalRows.Close()
		for goalRows.Next() {
			var columns string
			var sg = &StoryboardGoal{
				GoalID:    "",
				GoalName:  "",
				SortOrder: 0,
				Columns:   make([]*StoryboardColumn, 0),
			}
			if err := goalRows.Scan(&sg.GoalID, &sg.SortOrder, &sg.GoalName, &columns); err != nil {
				log.Println(err)
			} else {
				goalColumns := make([]*StoryboardColumn, 0)
				jsonErr := json.Unmarshal([]byte(columns), &goalColumns)
				if jsonErr != nil {
					log.Println(jsonErr)
				}
				sg.Columns = goalColumns
				goals = append(goals, sg)
			}
		}
	}

	return goals
}

/*
	Storyboard Column
*/

// CreateStoryboardColumn adds a new column to a Storyboard
func CreateStoryboardColumn(StoryboardID string, GoalID string, userID string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call create_storyboard_column($1, $2);`, StoryboardID, GoalID,
	); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

/*
	Storyboard Story
*/

// CreateStoryboardStory adds a new story to a Storyboard
func CreateStoryboardStory(StoryboardID string, GoalID string, ColumnID string, userID string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call create_storyboard_story($1, $2, $3);`, StoryboardID, GoalID, ColumnID,
	); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// ReviseStoryName updates the story name by ID
func ReviseStoryName(StoryboardID string, userID string, StoryID string, StoryName string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call update_story_name($1, $2);`,
		StoryID,
		StoryName,
	); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// ReviseStoryContent updates the story content by ID
func ReviseStoryContent(StoryboardID string, userID string, StoryID string, StoryContent string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call update_story_content($1, $2);`,
		StoryID,
		StoryContent,
	); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// ReviseStoryColor updates the story color by ID
func ReviseStoryColor(StoryboardID string, userID string, StoryID string, StoryColor string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call update_story_color($1, $2);`,
		StoryID,
		StoryColor,
	); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// MoveStoryboardStory moves the story by ID to Goal/Column by ID
func MoveStoryboardStory(StoryboardID string, userID string, StoryID string, GoalID string, ColumnID string, PlaceBefore string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call move_story($1, $2, $3, $4);`,
		StoryID,
		GoalID,
		ColumnID,
		PlaceBefore,
	); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// DeleteStoryboardStory removes a story from the current board by ID
func DeleteStoryboardStory(StoryboardID string, userID string, StoryID string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call delete_storyboard_story($1);`, StoryID); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// DeleteStoryboardColumn removes a column from the current board by ID
func DeleteStoryboardColumn(StoryboardID string, userID string, ColumnID string) ([]*StoryboardGoal, error) {
	err := ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := db.Exec(
		`call delete_storyboard_column($1);`, ColumnID); err != nil {
		log.Println(err)
	}

	goals := GetStoryboardGoals(StoryboardID)

	return goals, nil
}

/*
	User
*/

// GetUser gets a user from db by ID
func GetUser(UserID string) (*User, error) {
	var w User

	e := db.QueryRow(
		`SELECT * FROM get_user($1);`,
		UserID,
	).Scan(
		&w.UserID,
		&w.UserName,
		&w.UserEmail,
		&w.UserType,
	)
	if e != nil {
		log.Println(e)
		return nil, errors.New("User Not found")
	}

	return &w, nil
}

// AuthUser attempts to authenticate the user
func AuthUser(UserEmail string, UserPassword string) (*User, error) {
	var w User
	var passHash string

	e := db.QueryRow(
		`SELECT * FROM get_user_auth_by_email($1)`,
		UserEmail,
	).Scan(
		&w.UserID,
		&w.UserName,
		&w.UserEmail,
		&w.UserType,
		&passHash,
	)
	if e != nil {
		log.Println(e)
		return nil, errors.New("User Not found")
	}

	if ComparePasswords(passHash, []byte(UserPassword)) == false {
		return nil, errors.New("Password invalid")
	}

	return &w, nil
}

// CreateUserGuest adds a new user guest to the db
func CreateUserGuest(UserName string) (*User, error) {
	var UserID string
	e := db.QueryRow(`INSERT INTO users (name) VALUES ($1) RETURNING id`, UserName).Scan(&UserID)
	if e != nil {
		log.Println(e)
		return nil, errors.New("Unable to create new user")
	}

	return &User{UserID: UserID, UserName: UserName}, nil
}

// CreateUserRegistered adds a new user registered to the db
func CreateUserRegistered(UserName string, UserEmail string, UserPassword string, ActiveUserID string) (NewUser *User, VerifyID string, RegisterErr error) {
	hashedPassword, hashErr := HashAndSalt([]byte(UserPassword))
	if hashErr != nil {
		return nil, "", hashErr
	}

	var UserID string
	var verifyID string
	UserType := "REGISTERED"

	if ActiveUserID != "" {
		e := db.QueryRow(
			`SELECT userId, verifyId FROM register_user($1, $2, $3, $4, $5);`,
			ActiveUserID,
			UserName,
			UserEmail,
			hashedPassword,
			UserType,
		).Scan(&UserID, &verifyID)
		if e != nil {
			log.Println(e)
			return nil, "", errors.New("a user with that email already exists")
		}
	} else {
		e := db.QueryRow(
			`SELECT userId, verifyId FROM register_user($1, $2, $3, $4);`,
			UserName,
			UserEmail,
			hashedPassword,
			UserType,
		).Scan(&UserID, &verifyID)
		if e != nil {
			log.Println(e)
			return nil, "", errors.New("a user with that email already exists")
		}
	}

	return &User{UserID: UserID, UserName: UserName, UserEmail: UserEmail, UserType: UserType}, verifyID, nil
}

// UpdateUserProfile attempts to update the users profile
func UpdateUserProfile(UserID string, UserName string) error {
	if _, err := db.Exec(
		`UPDATE users SET name = $2 WHERE id = $1;`,
		UserID,
		UserName,
	); err != nil {
		log.Println(err)
		return errors.New("Error attempting to update users profile")
	}

	return nil
}

// UserResetRequest inserts a new user reset request
func UserResetRequest(UserEmail string) (resetID string, userName string, resetErr error) {
	var ResetID sql.NullString
	var UserID sql.NullString
	var UserName sql.NullString

	e := db.QueryRow(`
		SELECT resetId, userId, userName FROM insert_user_reset($1);
		`,
		UserEmail,
	).Scan(&ResetID, &UserID, &UserName)
	if e != nil {
		log.Println("Unable to reset user: ", e)
		// we don't want to alert the user that the email isn't valid
		return "", "", nil
	}

	return ResetID.String, UserName.String, nil
}

// UserResetPassword attempts to reset a users password
func UserResetPassword(ResetID string, UserPassword string) (userName string, userEmail string, resetErr error) {
	var UserName sql.NullString
	var UserEmail sql.NullString

	hashedPassword, hashErr := HashAndSalt([]byte(UserPassword))
	if hashErr != nil {
		return "", "", hashErr
	}

	userErr := db.QueryRow(`
		SELECT
			w.name, w.email
		FROM user_reset wr
		LEFT JOIN usersw ON w.id = wr.user_id
		WHERE wr.reset_id = $1;
		`,
		ResetID,
	).Scan(&UserName, &UserEmail)
	if userErr != nil {
		log.Println("Unable to get user for password reset confirmation email: ", userErr)
		return "", "", userErr
	}

	if _, err := db.Exec(
		`call reset_user_password($1, $2)`, ResetID, hashedPassword); err != nil {
		return "", "", err
	}

	return UserName.String, UserEmail.String, nil
}

// UserUpdatePassword attempts to update a users password
func UserUpdatePassword(UserID string, UserPassword string) (userName string, userEmail string, resetErr error) {
	var UserName sql.NullString
	var UserEmail sql.NullString

	userErr := db.QueryRow(`
		SELECT
			w.name, w.email
		FROM users w
		WHERE w.id = $1;
		`,
		UserID,
	).Scan(&UserName, &UserEmail)
	if userErr != nil {
		log.Println("Unable to get user for password update: ", userErr)
		return "", "", userErr
	}

	hashedPassword, hashErr := HashAndSalt([]byte(UserPassword))
	if hashErr != nil {
		return "", "", hashErr
	}

	if _, err := db.Exec(
		`call update_user_password($1, $2)`, UserID, hashedPassword); err != nil {
		return "", "", err
	}

	return UserName.String, UserEmail.String, nil
}

// VerifyUserAccount attempts to verify a users account email
func VerifyUserAccount(VerifyID string) error {
	if _, err := db.Exec(
		`call verify_user_account($1)`, VerifyID); err != nil {
		return err
	}

	return nil
}

/*
 Admin
*/

// ConfirmAdmin confirms whether the user is infact a ADMIN
func ConfirmAdmin(AdminID string) error {
	var userType string
	e := db.QueryRow("SELECT coalesce(type, '') FROM users WHERE id = $1;", AdminID).Scan(&userType)
	if e != nil {
		log.Println(e)
		return errors.New("could not find users type")
	}

	if userType != "ADMIN" {
		return errors.New(("not admin"))
	}

	return nil
}

// ApplicationStats includes user, storyboard counts
type ApplicationStats struct {
	RegisteredCount   int `json:"registeredUserCount"`
	UnregisteredCount int `json:"unregisteredUserCount"`
	StoryboardCount   int `json:"storyboardCount"`
}

// GetAppStats gets counts of users (registered and unregistered), and storyboards
func GetAppStats(AdminID string) (*ApplicationStats, error) {
	var Appstats ApplicationStats
	err := ConfirmAdmin(AdminID)
	if err != nil {
		log.Println("User isn't admin")
		return nil, errors.New("incorrect permissions")
	}

	statsErr := db.QueryRow(`
		SELECT
			unregistered_user_count,
			registered_user_count,
			storyboard_count
		FROM get_app_stats();
		`,
	).Scan(
		&Appstats.UnregisteredCount,
		&Appstats.RegisteredCount,
		&Appstats.StoryboardCount,
	)
	if statsErr != nil {
		log.Println("Unable to get application stats: ", statsErr)
		return nil, statsErr
	}

	return &Appstats, nil
}

// PromoteUser promotes a user to ADMIN type
func PromoteUser(AdminID string, UserID string) error {
	err := ConfirmAdmin(AdminID)
	if err != nil {
		return errors.New("incorrect permissions")
	}

	if _, err := db.Exec(
		`call promote_user($1);`,
		UserID,
	); err != nil {
		log.Println(err)
		return errors.New("error attempting to promote user to ADMIN")
	}

	return nil
}
