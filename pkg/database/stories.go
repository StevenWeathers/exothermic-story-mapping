package database

import (
	"errors"
	"log"
)

// CreateStoryboardStory adds a new story to a Storyboard
func (d *Database) CreateStoryboardStory(StoryboardID string, GoalID string, ColumnID string, userID string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call create_storyboard_story($1, $2, $3);`, StoryboardID, GoalID, ColumnID,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// ReviseStoryName updates the story name by ID
func (d *Database) ReviseStoryName(StoryboardID string, userID string, StoryID string, StoryName string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call update_story_name($1, $2);`,
		StoryID,
		StoryName,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// ReviseStoryContent updates the story content by ID
func (d *Database) ReviseStoryContent(StoryboardID string, userID string, StoryID string, StoryContent string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call update_story_content($1, $2);`,
		StoryID,
		StoryContent,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// ReviseStoryColor updates the story color by ID
func (d *Database) ReviseStoryColor(StoryboardID string, userID string, StoryID string, StoryColor string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call update_story_color($1, $2);`,
		StoryID,
		StoryColor,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// ReviseStoryPoints updates the story points by ID
func (d *Database) ReviseStoryPoints(StoryboardID string, userID string, StoryID string, Points int) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call update_story_points($1, $2);`,
		StoryID,
		Points,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// ReviseStoryClosed updates the story closed status by ID
func (d *Database) ReviseStoryClosed(StoryboardID string, userID string, StoryID string, Closed bool) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call update_story_closed($1, $2);`,
		StoryID,
		Closed,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// MoveStoryboardStory moves the story by ID to Goal/Column by ID
func (d *Database) MoveStoryboardStory(StoryboardID string, userID string, StoryID string, GoalID string, ColumnID string, PlaceBefore string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call move_story($1, $2, $3, $4);`,
		StoryID,
		GoalID,
		ColumnID,
		PlaceBefore,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// DeleteStoryboardStory removes a story from the current board by ID
func (d *Database) DeleteStoryboardStory(StoryboardID string, userID string, StoryID string) ([]*StoryboardGoal, error) {
	err := d.ConfirmOwner(StoryboardID, userID)
	if err != nil {
		return nil, errors.New("Incorrect permissions")
	}

	if _, err := d.db.Exec(
		`call delete_storyboard_story($1);`, StoryID); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// AddStoryComment adds a comment to a story
func (d *Database) AddStoryComment(StoryboardID string, UserID string, StoryID string, Comment string) ([]*StoryboardGoal, error) {
	if _, err := d.db.Exec(
		`call story_comment_add($1, $2, $3, $4);`,
		StoryboardID,
		StoryID,
		UserID,
		Comment,
	); err != nil {
		log.Println(err)
	}

	goals := d.GetStoryboardGoals(StoryboardID)

	return goals, nil
}

// // UpdateStoryComment updates a story comment
// func (d *Database) UpdateStoryComment(StoryboardID string, UserID string, commentID string, Comment string) ([]*StoryboardGoal, error) {
// 	if _, err := d.db.Exec(
// 		`call story_comment_edit($1, $2, $3, $4);`,
// 		StoryboardID,
// 		commentID,
// 		UserID,
// 		Comment,
// 	); err != nil {
// 		log.Println(err)
// 	}

// 	goals := d.GetStoryboardGoals(StoryboardID)

// 	return goals, nil
// }

// // DeleteStoryComment deletes a story comment
// func (d *Database) DeleteStoryComment(StoryboardID string, UserID string, commentID string) ([]*StoryboardGoal, error) {
// 	if _, err := d.db.Exec(
// 		`call story_comment_delete($1, $2, $3, $4);`,
// 		StoryboardID,
// 		commentID,
// 		UserID,
// 	); err != nil {
// 		log.Println(err)
// 	}

// 	goals := d.GetStoryboardGoals(StoryboardID)

// 	return goals, nil
// }
