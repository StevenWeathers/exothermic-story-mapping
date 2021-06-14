package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/StevenWeathers/exothermic-story-mapping/pkg/database"
	"github.com/gorilla/mux"
)

// handleGetTeamByUser gets an team with user role
func (s *server) handleGetTeamByUser() http.HandlerFunc {
	type TeamResponse struct {
		Team     *database.Team `json:"team"`
		TeamRole string         `json:"teamRole"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		TeamRole := r.Context().Value(contextKeyTeamRole).(string)
		TeamID := vars["teamId"]

		Team, err := s.database.TeamGet(TeamID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.respondWithJSON(w, http.StatusOK, &TeamResponse{
			Team:     Team,
			TeamRole: TeamRole,
		})
	}
}

// handleGetTeamsByUser gets a list of teams the user is apart of
func (s *server) handleGetTeamsByUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		UserID := r.Context().Value(contextKeyUserID).(string)
		vars := mux.Vars(r)
		Limit, _ := strconv.Atoi(vars["limit"])
		Offset, _ := strconv.Atoi(vars["offset"])

		Organizations := s.database.TeamListByUser(UserID, Limit, Offset)

		s.respondWithJSON(w, http.StatusOK, Organizations)
	}
}

// handleGetTeamUsers gets a list of users associated to the team
func (s *server) handleGetTeamUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		TeamID := vars["teamId"]
		Limit, _ := strconv.Atoi(vars["limit"])
		Offset, _ := strconv.Atoi(vars["offset"])

		Teams := s.database.TeamUserList(TeamID, Limit, Offset)

		s.respondWithJSON(w, http.StatusOK, Teams)
	}
}

// handleCreateTeam handles creating an team with current user as admin
func (s *server) handleCreateTeam() http.HandlerFunc {
	type CreateTeamResponse struct {
		TeamID string `json:"id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		UserID := r.Context().Value(contextKeyUserID).(string)
		keyVal := s.getJSONRequestBody(r, w)

		TeamName := keyVal["name"].(string)
		TeamID, err := s.database.TeamCreate(UserID, TeamName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var NewTeam = &CreateTeamResponse{
			TeamID: TeamID,
		}

		s.respondWithJSON(w, http.StatusOK, NewTeam)
	}
}

// handleTeamAddUser handles adding user to a team
func (s *server) handleTeamAddUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyVal := s.getJSONRequestBody(r, w)

		vars := mux.Vars(r)
		TeamID := vars["teamId"]
		UserEmail := strings.ToLower(keyVal["email"].(string))
		Role := keyVal["role"].(string)

		User, UserErr := s.database.GetUserByEmail(UserEmail)
		if UserErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err := s.database.TeamAddUser(TeamID, User.UserID, Role)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}

// handleTeamRemoveUser handles removing user from a team
func (s *server) handleTeamRemoveUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyVal := s.getJSONRequestBody(r, w)

		vars := mux.Vars(r)
		TeamID := vars["teamId"]
		UserID := keyVal["id"].(string)

		err := s.database.TeamRemoveUser(TeamID, UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}

// handleGetTeamStoryboards gets a list of storyboards associated to the team
func (s *server) handleGetTeamStoryboards() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		TeamID := vars["teamId"]
		Limit, _ := strconv.Atoi(vars["limit"])
		Offset, _ := strconv.Atoi(vars["offset"])

		Storyboards := s.database.TeamStoryboardList(TeamID, Limit, Offset)

		s.respondWithJSON(w, http.StatusOK, Storyboards)
	}
}

// handleTeamRemoveStoryboard handles removing storyboard from a team
func (s *server) handleTeamRemoveStoryboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyVal := s.getJSONRequestBody(r, w)

		vars := mux.Vars(r)
		TeamID := vars["teamId"]
		StoryboardID := keyVal["id"].(string)

		err := s.database.TeamRemoveStoryboard(TeamID, StoryboardID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}

// handleDeleteTeam handles deleting a team
func (s *server) handleDeleteTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyVal := s.getJSONRequestBody(r, w)
		TeamID := keyVal["id"].(string)

		err := s.database.TeamDelete(TeamID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}
