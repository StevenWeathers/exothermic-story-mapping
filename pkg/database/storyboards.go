package database

import (
	"encoding/json"
	"errors"
	"log"
)

//CreateStoryboard adds a new storyboard to the db
func (d *Database) CreateStoryboard(OwnerID string, StoryboardName string) (*Storyboard, error) {
	var b = &Storyboard{
		StoryboardID:   "",
		OwnerID:        OwnerID,
		StoryboardName: StoryboardName,
		Users:          make([]*StoryboardUser, 0),
	}

	e := d.db.QueryRow(
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
func (d *Database) GetStoryboard(StoryboardID string) (*Storyboard, error) {
	var cl string
	var b = &Storyboard{
		StoryboardID:   StoryboardID,
		OwnerID:        "",
		StoryboardName: "",
		Users:          make([]*StoryboardUser, 0),
		Goals:          make([]*StoryboardGoal, 0),
		ColorLegend:    make([]*Color, 0),
		Personas:       make([]*StoryboardPersona, 0),
	}

	// get storyboard
	e := d.db.QueryRow(
		`SELECT
			id, name, owner_id, color_legend
		FROM storyboard WHERE id = $1`,
		StoryboardID,
	).Scan(
		&b.StoryboardID,
		&b.StoryboardName,
		&b.OwnerID,
		&cl,
	)
	if e != nil {
		log.Println(e)
		return nil, errors.New("Not found")
	}

	clErr := json.Unmarshal([]byte(cl), &b.ColorLegend)
	if clErr != nil {
		log.Println(clErr)
	}

	b.Users = d.GetStoryboardUsers(StoryboardID)
	b.Goals = d.GetStoryboardGoals(StoryboardID)
	b.Personas = d.GetStoryboardPersonas(StoryboardID)

	return b, nil
}

// GetStoryboardsByUser gets a list of storyboards by UserID
func (d *Database) GetStoryboardsByUser(UserID string) ([]*Storyboard, error) {
	var storyboards = make([]*Storyboard, 0)
	storyboardRows, storyboardsErr := d.db.Query(`
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
			Users:          make([]*StoryboardUser, 0),
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
func (d *Database) ConfirmOwner(StoryboardID string, userID string) error {
	var ownerID string
	e := d.db.QueryRow("SELECT owner_id FROM storyboard WHERE id = $1", StoryboardID).Scan(&ownerID)
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
func (d *Database) GetStoryboardUser(StoryboardID string, UserID string) (*StoryboardUser, error) {
	var active bool
	var w StoryboardUser

	e := d.db.QueryRow(
		`SELECT * FROM get_storyboard_user($1, $2);`,
		StoryboardID,
		UserID,
	).Scan(
		&w.UserID,
		&w.UserName,
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
func (d *Database) GetStoryboardUsers(StoryboardID string) []*StoryboardUser {
	var users = make([]*StoryboardUser, 0)
	rows, err := d.db.Query(
		`SELECT * FROM get_storyboard_users($1);`,
		StoryboardID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var w StoryboardUser
			if err := rows.Scan(&w.UserID, &w.UserName, &w.Active); err != nil {
				log.Println(err)
			} else {
				users = append(users, &w)
			}
		}
	}

	return users
}

// GetStoryboardPersonas retrieves the personas for a given storyboard from db
func (d *Database) GetStoryboardPersonas(StoryboardID string) []*StoryboardPersona {
	var personas = make([]*StoryboardPersona, 0)
	rows, err := d.db.Query(
		`SELECT * FROM get_storyboard_personas($1);`,
		StoryboardID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p StoryboardPersona
			if err := rows.Scan(&p.PersonaID, &p.Name, &p.Role, &p.Description); err != nil {
				log.Println(err)
			} else {
				personas = append(personas, &p)
			}
		}
	}

	return personas
}

// AddUserToStoryboard adds a user by ID to the storyboard by ID
func (d *Database) AddUserToStoryboard(StoryboardID string, UserID string) ([]*StoryboardUser, error) {
	if _, err := d.db.Exec(
		`INSERT INTO storyboard_user (storyboard_id, user_id, active)
		VALUES ($1, $2, true)
		ON CONFLICT (storyboard_id, user_id) DO UPDATE SET active = true, abandoned = false`,
		StoryboardID,
		UserID,
	); err != nil {
		log.Println(err)
	}

	users := d.GetStoryboardUsers(StoryboardID)

	return users, nil
}

// RetreatUser removes a user from the current storyboard by ID
func (d *Database) RetreatUser(StoryboardID string, UserID string) []*StoryboardUser {
	if _, err := d.db.Exec(
		`UPDATE storyboard_user SET active = false WHERE storyboard_id = $1 AND user_id = $2`, StoryboardID, UserID); err != nil {
		log.Println(err)
	}

	if _, err := d.db.Exec(
		`UPDATE users SET last_active = NOW() WHERE id = $1`, UserID); err != nil {
		log.Println(err)
	}

	users := d.GetStoryboardUsers(StoryboardID)

	return users
}

// AbandonStoryboard removes a user from the current storyboard by ID and sets abandoned true
func (d *Database) AbandonStoryboard(StoryboardID string, UserID string) ([]*StoryboardUser, error) {
	if _, err := d.db.Exec(
		`UPDATE storyboard_user SET active = false, abandoned = true WHERE storyboard_id = $1 AND user_id = $2`, StoryboardID, UserID); err != nil {
		log.Println(err)
		return nil, err
	}

	if _, err := d.db.Exec(
		`UPDATE users SET last_active = NOW() WHERE id = $1`, UserID); err != nil {
		log.Println(err)
		return nil, err
	}

	users := d.GetStoryboardUsers(StoryboardID)

	return users, nil
}

// SetStoryboardOwner sets the ownerId for the storyboard
func (d *Database) SetStoryboardOwner(StoryboardID string, userID string, OwnerID string) (*Storyboard, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call set_storyboard_owner($1, $2);`, StoryboardID, OwnerID); err != nil {
		log.Println(err)
	}

	storyboard, err := d.GetStoryboard(StoryboardID)
	if err != nil {
		return nil, errors.New("Unable to promote owner")
	}

	return storyboard, nil
}

// ReviseColorLegend revises the storyboard color legend by StoryboardID
func (d *Database) ReviseColorLegend(StoryboardID string, UserID string, ColorLegend string) (*Storyboard, error) {
	err := d.ConfirmOwner(StoryboardID, UserID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call revise_color_legend($1, $2);`,
		StoryboardID,
		ColorLegend,
	); err != nil {
		log.Println(err)
		return nil, err
	}

	storyboard, err := d.GetStoryboard(StoryboardID)
	if err != nil {
		return nil, errors.New("Unable to promote owner")
	}

	return storyboard, nil
}

// DeleteStoryboard removes all storyboard associations and the storyboard itself from DB by StoryboardID
func (d *Database) DeleteStoryboard(StoryboardID string, userID string) error {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call delete_storyboard($1);`, StoryboardID); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// AddPersona adds a persona to a storyboard
func (d *Database) AddPersona(StoryboardID string, UserID string, Name string, Role string, Description string) ([]*StoryboardPersona, error) {
	err := d.ConfirmOwner(StoryboardID, UserID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call persona_add($1, $2, $3, $4);`,
		StoryboardID,
		Name,
		Role,
		Description,
	); err != nil {
		log.Println(err)
	}

	personas := d.GetStoryboardPersonas(StoryboardID)

	return personas, nil
}

// UpdatePersona updates a storyboard persona
func (d *Database) UpdatePersona(StoryboardID string, UserID string, PersonaID string, Name string, Role string, Description string) ([]*StoryboardPersona, error) {
	err := d.ConfirmOwner(StoryboardID, UserID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call persona_edit($1, $2, $3, $4, $5);`,
		StoryboardID,
		PersonaID,
		Name,
		Role,
		Description,
	); err != nil {
		log.Println(err)
	}

	personas := d.GetStoryboardPersonas(StoryboardID)

	return personas, nil
}

// DeletePersona deletes a storyboard persona
func (d *Database) DeletePersona(StoryboardID string, UserID string, PersonaID string) ([]*StoryboardPersona, error) {
	err := d.ConfirmOwner(StoryboardID, UserID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call persona_delete($1, $2);`,
		StoryboardID,
		PersonaID,
	); err != nil {
		log.Println(err)
	}

	personas := d.GetStoryboardPersonas(StoryboardID)

	return personas, nil
}
