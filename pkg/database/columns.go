package database

import (
	"errors"
	"log"
)

// CreateStoryboardColumn adds a new column to a Storyboard
func (d *Database) CreateStoryboardColumn(StoryboardID string, GoalID string, userID string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call create_storyboard_column($1, $2);`, StoryboardID, GoalID,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// DeleteStoryboardColumn removes a column from the current board by ID
func (d *Database) DeleteStoryboardColumn(StoryboardID string, userID string, ColumnID string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call delete_storyboard_column($1);`, ColumnID); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}
