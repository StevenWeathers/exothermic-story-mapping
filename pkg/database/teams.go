package database

import (
	"errors"
	"log"
)

// TeamUserRole gets a users role in team
func (d *Database) TeamUserRole(UserID string, TeamID string) (string, error) {
	var teamRole string

	e := d.db.QueryRow(
		`SELECT role FROM team_get_user_role($1, $2)`,
		UserID,
		TeamID,
	).Scan(
		&teamRole,
	)
	if e != nil {
		log.Println(e)
		return "", errors.New("error getting team users role")
	}

	return teamRole, nil
}

// TeamGet gets an team
func (d *Database) TeamGet(TeamID string) (*Team, error) {
	var team = &Team{
		TeamID:      "",
		Name:        "",
		CreatedDate: "",
		UpdatedDate: "",
	}

	e := d.db.QueryRow(
		`SELECT id, name, created_date, updated_date FROM team_get_by_id($1)`,
		TeamID,
	).Scan(
		&team.TeamID,
		&team.Name,
		&team.CreatedDate,
		&team.UpdatedDate,
	)
	if e != nil {
		log.Println(e)
		return nil, errors.New("team not found")
	}

	return team, nil
}

// TeamListByUser gets a list of teams the user is on
func (d *Database) TeamListByUser(UserID string, Limit int, Offset int) []*Team {
	var teams = make([]*Team, 0)
	rows, err := d.db.Query(
		`SELECT id, name, created_date, updated_date FROM team_list_by_user($1, $2, $3);`,
		UserID,
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

// TTeamCreate creates a team with current user as an ADMIN
func (d *Database) TeamCreate(UserID string, TeamName string) (string, error) {
	var TeamID string
	err := d.db.QueryRow(`
		SELECT teamId FROM team_create($1, $2);`,
		UserID,
		TeamName,
	).Scan(&TeamID)

	if err != nil {
		log.Println("Unable to create team: ", err)
		return "", err
	}

	return TeamID, nil
}

// TeamAddUser adds a user to a team
func (d *Database) TeamAddUser(TeamID string, UserID string, Role string) (string, error) {
	_, err := d.db.Exec(
		`SELECT team_user_add($1, $2, $3);`,
		TeamID,
		UserID,
		Role,
	)

	if err != nil {
		log.Println("Unable to add user to team: ", err)
		return "", err
	}

	return TeamID, nil
}

// TeamUserList gets a list of team users
func (d *Database) TeamUserList(TeamID string, Limit int, Offset int) []*OrganizationUser {
	var users = make([]*OrganizationUser, 0)
	rows, err := d.db.Query(
		`SELECT id, name, email, role FROM team_user_list($1, $2, $3);`,
		TeamID,
		Limit,
		Offset,
	)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var usr OrganizationUser

			if err := rows.Scan(
				&usr.UserID,
				&usr.Name,
				&usr.Email,
				&usr.Role,
			); err != nil {
				log.Println(err)
			} else {
				users = append(users, &usr)
			}
		}
	} else {
		log.Println(err)
	}

	return users
}

// TeamRemoveUser removes a user from a team
func (d *Database) TeamRemoveUser(TeamID string, UserID string) error {
	_, err := d.db.Exec(
		`CALL team_user_remove($1, $2);`,
		TeamID,
		UserID,
	)

	if err != nil {
		log.Println("Unable to remove user from team: ", err)
		return err
	}

	return nil
}

// TeamStoryboardList gets a list of team storyboards
func (d *Database) TeamStoryboardList(TeamID string, Limit int, Offset int) []*Storyboard {
	var storyboards = make([]*Storyboard, 0)
	rows, err := d.db.Query(
		`SELECT id, name FROM team_storyboard_list($1, $2, $3);`,
		TeamID,
		Limit,
		Offset,
	)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tb Storyboard

			if err := rows.Scan(
				&tb.StoryboardID,
				&tb.StoryboardName,
			); err != nil {
				log.Println(err)
			} else {
				storyboards = append(storyboards, &tb)
			}
		}
	} else {
		log.Println(err)
	}

	return storyboards
}

// TeamAddStoryboard adds a storyboard to a team
func (d *Database) TeamAddStoryboard(TeamID string, StoryboardID string) error {
	_, err := d.db.Exec(
		`SELECT team_storyboard_add($1, $2);`,
		TeamID,
		StoryboardID,
	)

	if err != nil {
		log.Println("Unable to add storyboard to team: ", err)
		return err
	}

	return nil
}

// TeamRemoveStoryboard removes a storyboard from a team
func (d *Database) TeamRemoveStoryboard(TeamID string, StoryboardID string) error {
	_, err := d.db.Exec(
		`SELECT team_storyboard_remove($1, $2);`,
		TeamID,
		StoryboardID,
	)

	if err != nil {
		log.Println("Unable to remove storyboard from team: ", err)
		return err
	}

	return nil
}

// TeamDelete deletes a team
func (d *Database) TeamDelete(TeamID string) error {
	_, err := d.db.Exec(
		`CALL team_delete($1);`,
		TeamID,
	)

	if err != nil {
		log.Println("Unable to delete team: ", err)
		return err
	}

	return nil
}
