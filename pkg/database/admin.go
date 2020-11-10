package database

import (
	"errors"
	"log"
)

// ConfirmAdmin confirms whether the user is infact a ADMIN
func (d *Database) ConfirmAdmin(AdminID string) error {
	var userType string
	e := d.db.QueryRow("SELECT coalesce(type, '') FROM users WHERE id = $1;", AdminID).Scan(&userType)
	if e != nil {
		log.Println(e)
		return errors.New("could not find users type")
	}

	if userType != "ADMIN" {
		return errors.New(("user is not an admin"))
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
func (d *Database) GetAppStats() (*ApplicationStats, error) {
	var Appstats ApplicationStats

	statsErr := d.db.QueryRow(`
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
func (d *Database) PromoteUser(UserID string) error {
	if _, err := d.db.Exec(
		`call promote_user($1);`,
		UserID,
	); err != nil {
		log.Println(err)
		return errors.New("error attempting to promote user to ADMIN")
	}

	return nil
}

// DemoteUser demotes a user to REGISTERED type
func (d *Database) DemoteUser(UserID string) error {
	if _, err := d.db.Exec(
		`call demote_user($1);`,
		UserID,
	); err != nil {
		log.Println(err)
		return errors.New("error attempting to demote user to REGISTERED")
	}

	return nil
}
