package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/lib/pq"
	"github.com/markbates/pkger"
)

var db *sql.DB

// Storyboard aka arena
type Storyboard struct {
	StoryboardID   string  `json:"id"`
	OwnerID        string  `json:"ownerId"`
	StoryboardName string  `json:"name"`
	Users          []*User `json:"users"`
}

// User aka user
type User struct {
	UserID    string `json:"id"`
	UserName  string `json:"name"`
	UserEmail string `json:"email"`
	UserType  string `json:"type"`
	Active    bool   `json:"active"`
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
		`INSERT INTO storyboard (owner_id, name) VALUES ($1, $2) RETURNING id`,
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

	return b, nil
}

// GetStoryboardsByUser gets a list of storyboards by UserID
func GetStoryboardsByUser(UserID string) ([]*Storyboard, error) {
	var storyboards = make([]*Storyboard, 0)
	storyboardRows, storyboardsErr := db.Query(`
		SELECT b.id, b.name, b.owner_id
		FROM storyboard b
		LEFT JOIN storyboard_user bw ON b.id = bw.storyboard_id WHERE bw.user_id = $1
		GROUP BY b.id ORDER BY b.created_date DESC
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
		`SELECT
			w.id, w.name, coalesce(w.email, ''), w.type, coalesce(bw.active, FALSE)
		FROM users w
		LEFT JOIN storyboard_user bw ON bw.user_id = w.id AND bw.storyboard_id = $1
		WHERE id = $2`,
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
		`SELECT
			w.id, w.name, w.email, w.type, bw.active
		FROM storyboard_user bw
		LEFT JOIN users w ON bw.user_id = w.id
		WHERE bw.storyboard_id = $1
		ORDER BY w.name`,
		StoryboardID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var w User
			var userEmail sql.NullString
			if err := rows.Scan(&w.UserID, &w.UserName, &userEmail, &w.UserType, &w.Active); err != nil {
				log.Println(err)
			} else {
				w.UserEmail = userEmail.String
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

	// set storyboard VotingLocked
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
	User
*/

// GetUser gets a user from db by ID
func GetUser(UserID string) (*User, error) {
	var w User
	var userEmail sql.NullString

	e := db.QueryRow(
		"SELECT id, name, email, type FROM users WHERE id = $1",
		UserID,
	).Scan(
		&w.UserID,
		&w.UserName,
		&userEmail,
		&w.UserType,
	)
	if e != nil {
		log.Println(e)
		return nil, errors.New("User Not found")
	}

	w.UserEmail = userEmail.String

	return &w, nil
}

// AuthUser attempts to authenticate the user
func AuthUser(UserEmail string, UserPassword string) (*User, error) {
	var w User
	var passHash string

	e := db.QueryRow(
		`SELECT id, name, email, type, password FROM users WHERE email = $1`,
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
func CreateUserRegistered(UserName string, UserEmail string, UserPassword string) (*User, error) {
	hashedPassword, hashErr := HashAndSalt([]byte(UserPassword))
	if hashErr != nil {
		return nil, hashErr
	}

	var UserID string
	UserType := "REGISTERED"

	e := db.QueryRow(
		`INSERT INTO users (name, email, password, type) VALUES ($1, $2, $3, $4) RETURNING id`,
		UserName,
		UserEmail,
		hashedPassword,
		UserType,
	).Scan(&UserID)
	if e != nil {
		log.Println(e)
		return nil, errors.New("a User with that email already exists")
	}

	return &User{UserID: UserID, UserName: UserName, UserEmail: UserEmail, UserType: UserType}, nil
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

	warErr := db.QueryRow(`
		SELECT w.id, w.name FROM users w WHERE w.email = $1;
		`,
		UserEmail,
	).Scan(&UserID, &UserName)
	if warErr != nil {
		log.Println("Unable to get user for reset email: ", warErr)
		// we don't want to alert the user that the email isn't valid
		return "", "", nil
	}

	e := db.QueryRow(`
		INSERT INTO user_reset (user_id)
		VALUES ($1)
		RETURNING reset_id;
		`,
		UserID.String,
	).Scan(&ResetID)
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

	warErr := db.QueryRow(`
		SELECT
			w.name, w.email
		FROM user_reset wr
		LEFT JOIN usersw ON w.id = wr.user_id
		WHERE wr.reset_id = $1;
		`,
		ResetID,
	).Scan(&UserName, &UserEmail)
	if warErr != nil {
		log.Println("Unable to get user for password reset confirmation email: ", warErr)
		return "", "", warErr
	}

	if _, err := db.Exec(
		`call reset_user_password($1, $2)`, ResetID, hashedPassword); err != nil {
		return "", "", err
	}

	return UserName.String, UserEmail.String, nil
}
