package database

import (
	"encoding/json"
	"errors"
	"log"
)

// CreateStoryboardGoal adds a new goal to a Storyboard
func (d *Database) CreateStoryboardGoal(StoryboardID string, userID string, GoalName string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call create_storyboard_goal($1, $2);`, StoryboardID, GoalName,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// ReviseGoalName updates the plan name by ID
func (d *Database) ReviseGoalName(StoryboardID string, userID string, GoalID string, GoalName string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call update_storyboard_goal($1, $2);`,
		GoalID,
		GoalName,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// DeleteStoryboardGoal removes a goal from the current board by ID
func (d *Database) DeleteStoryboardGoal(StoryboardID string, userID string, GoalID string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call delete_storyboard_goal($1);`, GoalID); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// GetStoryboardGoals retrieves goals for given storyboard from db
func (d *Database) GetStoryboardGoals(StoryboardID string) []*StoryboardGoal {
	var goals = make([]*StoryboardGoal, 0)

	goalRows, goalsErr := d.db.Query(
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
