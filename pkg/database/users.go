package database

import (
	"database/sql"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// HashAndSalt takes a password byte and salt + hashes it
// returning a hash string to store in db
func HashAndSalt(pwd []byte) (string, error) {
	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
		return "", err
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash), nil
}

// ComparePasswords takes a password hash and compares it to entered password bytes
// returning true if matches false if not
func ComparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

// GetRegisteredUsers retrieves the registered users from db
func (d *Database) GetRegisteredUsers() []*User {
	var users = make([]*User, 0)
	rows, err := d.db.Query(
		"SELECT id, name, email, type, avatar, verified FROM users WHERE email IS NOT NULL",
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var w User
			var userEmail sql.NullString

			if err := rows.Scan(&w.UserID,
				&w.UserName,
				&userEmail,
				&w.UserType,
				&w.UserAvatar,
				&w.Verified,
			); err != nil {
				log.Println(err)
			} else {
				w.UserEmail = userEmail.String
				users = append(users, &w)
			}
		}
	}

	return users
}

// GetUser gets a user from db by ID
func (d *Database) GetUser(UserID string) (*User, error) {
	var w User

	e := d.db.QueryRow(
		`SELECT * FROM get_user($1);`,
		UserID,
	).Scan(
		&w.UserID,
		&w.UserName,
		&w.UserEmail,
		&w.UserType,
		&w.Verified,
	)
	if e != nil {
		log.Println(e)
		return nil, errors.New("User Not found")
	}

	return &w, nil
}

// GetUserByEmail gets a user by email
func (d *Database) GetUserByEmail(UserEmail string) (*User, error) {
	var u User
	e := d.db.QueryRow(
		"SELECT id, name, email, type, verified FROM users WHERE email = $1",
		UserEmail,
	).Scan(
		&u.UserID,
		&u.UserName,
		&u.UserEmail,
		&u.UserType,
		&u.Verified,
	)
	if e != nil {
		log.Println(e)
		return nil, errors.New("user email not found")
	}

	return &u, nil
}

// AuthUser attempts to authenticate the user
func (d *Database) AuthUser(UserEmail string, UserPassword string) (*User, error) {
	var w User
	var passHash string

	e := d.db.QueryRow(
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
func (d *Database) CreateUserGuest(UserName string) (*User, error) {
	var UserID string
	e := d.db.QueryRow(`INSERT INTO users (name) VALUES ($1) RETURNING id`, UserName).Scan(&UserID)
	if e != nil {
		log.Println(e)
		return nil, errors.New("Unable to create new user")
	}

	return &User{UserID: UserID, UserName: UserName}, nil
}

// CreateUserRegistered adds a new user registered to the db
func (d *Database) CreateUserRegistered(UserName string, UserEmail string, UserPassword string, ActiveUserID string) (NewUser *User, VerifyID string, RegisterErr error) {
	hashedPassword, hashErr := HashAndSalt([]byte(UserPassword))
	if hashErr != nil {
		return nil, "", hashErr
	}

	var UserID string
	var verifyID string
	UserType := "REGISTERED"

	if ActiveUserID != "" {
		e := d.db.QueryRow(
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
		e := d.db.QueryRow(
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
func (d *Database) UpdateUserProfile(UserID string, UserName string, UserAvatar string) error {
	if _, err := d.db.Exec(
		`UPDATE users SET name = $2, avatar = $3 WHERE id = $1;`,
		UserID,
		UserName,
		UserAvatar,
	); err != nil {
		log.Println(err)
		return errors.New("Error attempting to update users profile")
	}

	return nil
}

// UserResetRequest inserts a new user reset request
func (d *Database) UserResetRequest(UserEmail string) (resetID string, userName string, resetErr error) {
	var ResetID sql.NullString
	var UserID sql.NullString
	var UserName sql.NullString

	e := d.db.QueryRow(`
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
func (d *Database) UserResetPassword(ResetID string, UserPassword string) (userName string, userEmail string, resetErr error) {
	var UserName sql.NullString
	var UserEmail sql.NullString

	hashedPassword, hashErr := HashAndSalt([]byte(UserPassword))
	if hashErr != nil {
		return "", "", hashErr
	}

	userErr := d.db.QueryRow(`
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

	if _, err := d.db.Exec(
		`call reset_user_password($1, $2)`, ResetID, hashedPassword); err != nil {
		return "", "", err
	}

	return UserName.String, UserEmail.String, nil
}

// UserUpdatePassword attempts to update a users password
func (d *Database) UserUpdatePassword(UserID string, UserPassword string) (userName string, userEmail string, resetErr error) {
	var UserName sql.NullString
	var UserEmail sql.NullString

	userErr := d.db.QueryRow(`
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

	if _, err := d.db.Exec(
		`call update_user_password($1, $2)`, UserID, hashedPassword); err != nil {
		return "", "", err
	}

	return UserName.String, UserEmail.String, nil
}

// VerifyUserAccount attempts to verify a users account email
func (d *Database) VerifyUserAccount(VerifyID string) error {
	if _, err := d.db.Exec(
		`call verify_user_account($1)`, VerifyID); err != nil {
		return err
	}

	return nil
}
