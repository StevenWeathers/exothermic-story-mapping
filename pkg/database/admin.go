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

// GetAppStats gets counts of users (registered and unregistered), and storyboards
func (d *Database) GetAppStats() (*ApplicationStats, error) {
	var Appstats ApplicationStats

	statsErr := d.db.QueryRow(`
		SELECT
			unregistered_user_count,
			registered_user_count,
			storyboard_count,
			organization_count,
			department_count,
			team_count
		FROM get_app_stats();
		`,
	).Scan(
		&Appstats.UnregisteredCount,
		&Appstats.RegisteredCount,
		&Appstats.StoryboardCount,
		&Appstats.OrganizationCount,
		&Appstats.DepartmentCount,
		&Appstats.TeamCount,
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

// CleanStoryboards deletes storyboards older than X days
func (d *Database) CleanStoryboards(DaysOld int) error {
	if _, err := d.db.Exec(
		`call clean_storyboards($1);`,
		DaysOld,
	); err != nil {
		log.Println(err)
		return errors.New("error attempting to clean storyboards")
	}

	return nil
}

// CleanGuests deletes guest users older than X days
func (d *Database) CleanGuests(DaysOld int) error {
	if _, err := d.db.Exec(
		`call clean_guest_users($1);`,
		DaysOld,
	); err != nil {
		log.Println(err)
		return errors.New("error attempting to clean Guest Warriors")
	}

	return nil
}

// OrganizationList gets a list of organizations
func (d *Database) OrganizationList(Limit int, Offset int) []*Organization {
	var organizations = make([]*Organization, 0)
	rows, err := d.db.Query(
		`SELECT id, name, created_date, updated_date FROM organization_list($1, $2);`,
		Limit,
		Offset,
	)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var org Organization

			if err := rows.Scan(
				&org.OrganizationID,
				&org.Name,
				&org.CreatedDate,
				&org.UpdatedDate,
			); err != nil {
				log.Println(err)
			} else {
				organizations = append(organizations, &org)
			}
		}
	} else {
		log.Println(err)
	}

	return organizations
}

// TeamList gets a list of teams
func (d *Database) TeamList(Limit int, Offset int) []*Team {
	var teams = make([]*Team, 0)
	rows, err := d.db.Query(
		`SELECT id, name, created_date, updated_date FROM team_list($1, $2);`,
		Limit,
		Offset,
	)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var team Team

			if err := rows.Scan(
				&team.TeamID,
				&team.Name,
				&team.CreatedDate,
				&team.UpdatedDate,
			); err != nil {
				log.Println(err)
			} else {
				teams = append(teams, &team)
			}
		}
	} else {
		log.Println(err)
	}

	return teams
}
