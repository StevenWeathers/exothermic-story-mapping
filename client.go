package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// SocketEvent is the event structure used for socket messages
type SocketEvent struct {
	EventType  string `json:"type"`
	EventValue string `json:"value"`
	EventUser  string `json:"userId"`
}

// CreateSocketEvent makes a SocketEvent struct and turns it into json []byte
func CreateSocketEvent(EventType string, EventValue string, EventUser string) []byte {
	newEvent := &SocketEvent{
		EventType:  EventType,
		EventValue: EventValue,
		EventUser:  EventUser,
	}

	event, _ := json.Marshal(newEvent)

	return event
}

// readPump pumps messages from the websocket connection to the hub.
func (s subscription) readPump(srv *server) {
	var forceClosed bool
	c := s.conn
	defer func() {
		StoryboardID := s.arena
		UserID := s.userID

		Users := srv.database.RetreatUser(StoryboardID, UserID)
		updatedUsers, _ := json.Marshal(Users)

		retreatEvent := CreateSocketEvent("user_retreated", string(updatedUsers), UserID)
		m := message{retreatEvent, StoryboardID}
		h.broadcast <- m

		h.unregister <- s
		if forceClosed {
			cm := websocket.FormatCloseMessage(4002, "abandoned")
			if err := c.ws.WriteControl(websocket.CloseMessage, cm, time.Now().Add(writeWait)); err != nil {
				log.Printf("abandon error: %v", err)
			}
		}
		if err := c.ws.Close(); err != nil {
			log.Printf("close error: %v", err)
		}
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}

		var badEvent bool
		keyVal := make(map[string]string)
		json.Unmarshal(msg, &keyVal) // check for errors
		userID := s.userID
		storyboardID := s.arena

		switch keyVal["type"] {
		case "add_goal":
			goals, err := srv.database.CreateStoryboardGoal(storyboardID, userID, keyVal["value"])
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("goal_added", string(updatedGoals), "")
		case "revise_goal":
			goalObj := make(map[string]string)
			json.Unmarshal([]byte(keyVal["value"]), &goalObj)
			GoalID := goalObj["goalId"]
			GoalName := goalObj["name"]

			goals, err := srv.database.ReviseGoalName(storyboardID, userID, GoalID, GoalName)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("goal_revised", string(updatedGoals), "")
		case "delete_goal":
			goals, err := srv.database.DeleteStoryboardGoal(storyboardID, userID, keyVal["value"])
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("goal_deleted", string(updatedGoals), "")
		case "add_column":
			goalObj := make(map[string]string)
			json.Unmarshal([]byte(keyVal["value"]), &goalObj)
			GoalID := goalObj["goalId"]

			goals, err := srv.database.CreateStoryboardColumn(storyboardID, GoalID, userID)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("column_added", string(updatedGoals), "")
		case "revise_column":
			var rs struct {
				ColumnID string `json:"id"`
				Name     string `json:"name"`
			}
			json.Unmarshal([]byte(keyVal["value"]), &rs)

			goals, err := srv.database.ReviseStoryboardColumn(storyboardID, userID, rs.ColumnID, rs.Name)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("column_updated", string(updatedGoals), "")
		case "delete_column":
			goals, err := srv.database.DeleteStoryboardColumn(storyboardID, userID, keyVal["value"])
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_deleted", string(updatedGoals), "")
		case "add_story":
			goalObj := make(map[string]string)
			json.Unmarshal([]byte(keyVal["value"]), &goalObj)
			GoalID := goalObj["goalId"]
			ColumnID := goalObj["columnId"]

			goals, err := srv.database.CreateStoryboardStory(storyboardID, GoalID, ColumnID, userID)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_added", string(updatedGoals), "")
		case "update_story_name":
			goalObj := make(map[string]string)
			json.Unmarshal([]byte(keyVal["value"]), &goalObj)
			StoryID := goalObj["storyId"]
			StoryName := goalObj["name"]

			goals, err := srv.database.ReviseStoryName(storyboardID, userID, StoryID, StoryName)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_updated", string(updatedGoals), "")
		case "update_story_content":
			goalObj := make(map[string]string)
			json.Unmarshal([]byte(keyVal["value"]), &goalObj)
			StoryID := goalObj["storyId"]
			StoryContent := goalObj["content"]

			goals, err := srv.database.ReviseStoryContent(storyboardID, userID, StoryID, StoryContent)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_updated", string(updatedGoals), "")
		case "update_story_color":
			goalObj := make(map[string]string)
			json.Unmarshal([]byte(keyVal["value"]), &goalObj)
			StoryID := goalObj["storyId"]
			StoryColor := goalObj["color"]

			goals, err := srv.database.ReviseStoryColor(storyboardID, userID, StoryID, StoryColor)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_updated", string(updatedGoals), "")
		case "update_story_points":
			var rs struct {
				StoryID string `json:"storyId"`
				Points  int    `json:"points"`
			}
			json.Unmarshal([]byte(keyVal["value"]), &rs)

			goals, err := srv.database.ReviseStoryPoints(storyboardID, userID, rs.StoryID, rs.Points)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_updated", string(updatedGoals), "")
		case "update_story_closed":
			var rs struct {
				StoryID string `json:"storyId"`
				Closed  bool   `json:"closed"`
			}
			json.Unmarshal([]byte(keyVal["value"]), &rs)

			goals, err := srv.database.ReviseStoryClosed(storyboardID, userID, rs.StoryID, rs.Closed)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_updated", string(updatedGoals), "")
		case "move_story":
			goalObj := make(map[string]string)
			json.Unmarshal([]byte(keyVal["value"]), &goalObj)
			StoryID := goalObj["storyId"]
			GoalID := goalObj["goalId"]
			ColumnID := goalObj["columnId"]
			PlaceBefore := goalObj["placeBefore"]

			goals, err := srv.database.MoveStoryboardStory(storyboardID, userID, StoryID, GoalID, ColumnID, PlaceBefore)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_moved", string(updatedGoals), "")
		case "delete_story":
			goals, err := srv.database.DeleteStoryboardStory(storyboardID, userID, keyVal["value"])
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_deleted", string(updatedGoals), "")
		case "add_story_comment":
			var rs struct {
				StoryID string `json:"storyId"`
				Comment string `json:"comment"`
			}
			json.Unmarshal([]byte(keyVal["value"]), &rs)

			goals, err := srv.database.AddStoryComment(storyboardID, userID, rs.StoryID, rs.Comment)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_updated", string(updatedGoals), "")
		// case "update_story_comment":
		// 	var rs struct {
		// 		StoryID   string `json:"storyId"`
		// 		CommendID string `json:"commentId"`
		// 		Comment   string `json:"comment"`
		// 	}
		// 	json.Unmarshal([]byte(keyVal["value"]), &rs)

		// 	goals, err := srv.database.UpdateStoryComment(storyboardID, userID, rs.CommendID, rs.Comment)
		// 	if err != nil {
		// 		badEvent = true
		// 		break
		// 	}
		// 	updatedGoals, _ := json.Marshal(goals)
		// 	msg = CreateSocketEvent("story_updated", string(updatedGoals), "")
		// case "delete_story_comment":
		// 	var rs struct {
		// 		StoryID   string `json:"storyId"`
		// 		CommendID string `json:"commentId"`
		// 	}
		// 	json.Unmarshal([]byte(keyVal["value"]), &rs)

		// 	goals, err := srv.database.DeleteStoryComment(storyboardID, userID, rs.CommendID)
		// 	if err != nil {
		// 		badEvent = true
		// 		break
		// 	}
		// 	updatedGoals, _ := json.Marshal(goals)
		// 	msg = CreateSocketEvent("story_updated", string(updatedGoals), "")
		case "add_persona":
			var rs struct {
				Name        string `json:"name"`
				Role        string `json:"role"`
				Description string `json:"description"`
			}
			json.Unmarshal([]byte(keyVal["value"]), &rs)

			personas, err := srv.database.AddPersona(storyboardID, userID, rs.Name, rs.Role, rs.Description)
			if err != nil {
				badEvent = true
				break
			}
			updatedPersonas, _ := json.Marshal(personas)
			msg = CreateSocketEvent("personas_updated", string(updatedPersonas), "")
		case "update_persona":
			var rs struct {
				PersonaID   string `json:"id"`
				Name        string `json:"name"`
				Role        string `json:"role"`
				Description string `json:"description"`
			}
			json.Unmarshal([]byte(keyVal["value"]), &rs)

			personas, err := srv.database.UpdatePersona(storyboardID, userID, rs.PersonaID, rs.Name, rs.Role, rs.Description)
			if err != nil {
				badEvent = true
				break
			}
			updatedPersonas, _ := json.Marshal(personas)
			msg = CreateSocketEvent("personas_updated", string(updatedPersonas), "")
		case "delete_persona":
			personas, err := srv.database.DeletePersona(storyboardID, userID, keyVal["value"])
			if err != nil {
				badEvent = true
				break
			}
			updatedPersonas, _ := json.Marshal(personas)
			msg = CreateSocketEvent("personas_updated", string(updatedPersonas), "")
		case "promote_owner":
			storyboard, err := srv.database.SetStoryboardOwner(storyboardID, userID, keyVal["value"])
			if err != nil {
				badEvent = true
				break
			}

			updatedStoryboard, _ := json.Marshal(storyboard)
			msg = CreateSocketEvent("storyboard_updated", string(updatedStoryboard), "")
		case "revise_color_legend":
			storyboard, err := srv.database.ReviseColorLegend(storyboardID, userID, keyVal["value"])
			if err != nil {
				badEvent = true
				break
			}

			updatedStoryboard, _ := json.Marshal(storyboard)
			msg = CreateSocketEvent("storyboard_updated", string(updatedStoryboard), "")
		case "concede_storyboard":
			err := srv.database.DeleteStoryboard(storyboardID, userID)
			if err != nil {
				badEvent = true
				break
			}
			msg = CreateSocketEvent("storyboard_conceded", "", "")
		case "abandon_storyboard":
			_, err := srv.database.AbandonStoryboard(storyboardID, userID)
			if err != nil {
				badEvent = true
				break
			}
			badEvent = true // don't want this event to cause write panic
			forceClosed = true
		default:
		}

		if !badEvent {
			m := message{msg, s.arena}
			h.broadcast <- m
		}

		if forceClosed {
			break
		}
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (s *subscription) writePump() {
	c := s.conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func (s *server) serveWs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		storyboardID := vars["id"]

		// upgrade to WebSocket connection
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		// make sure user cookies are valid
		userID, cookieErr := s.validateUserCookie(w, r)
		if cookieErr != nil {
			cm := websocket.FormatCloseMessage(4001, "unauthorized")
			if err := ws.WriteMessage(websocket.CloseMessage, cm); err != nil {
				log.Printf("unauthorized close error: %v", err)
			}
			if err := ws.Close(); err != nil {
				log.Printf("close error: %v", err)
			}
			return
		}

		// make sure storyboard is legit
		b, storyboardErr := s.database.GetStoryboard(storyboardID)
		if storyboardErr != nil {
			cm := websocket.FormatCloseMessage(4004, "storyboard not found")
			if err := ws.WriteMessage(websocket.CloseMessage, cm); err != nil {
				log.Printf("not found close error: %v", err)
			}
			if err := ws.Close(); err != nil {
				log.Printf("close error: %v", err)
			}
			return
		}
		storyboard, _ := json.Marshal(b)

		// make sure user exists
		_, userErr := s.database.GetStoryboardUser(storyboardID, userID)

		if userErr != nil {
			log.Println("error finding user : " + userErr.Error() + "\n")
			cm := websocket.FormatCloseMessage(4003, "duplicate session")

			if fmt.Sprint(userErr) == "User Not found" {
				s.clearUserCookies(w)
				cm = websocket.FormatCloseMessage(4001, "unauthorized")
			}

			if err := ws.WriteMessage(websocket.CloseMessage, cm); err != nil {
				log.Printf("unauthorized close error: %v", err)
			}
			if err := ws.Close(); err != nil {
				log.Printf("close error: %v", err)
			}
			return
		}

		c := &connection{send: make(chan []byte, 256), ws: ws}
		ss := subscription{c, storyboardID, userID}
		h.register <- ss

		Users, _ := s.database.AddUserToStoryboard(ss.arena, userID)
		updatedUsers, _ := json.Marshal(Users)

		initEvent := CreateSocketEvent("init", string(storyboard), userID)
		_ = c.write(websocket.TextMessage, initEvent)

		joinedEvent := CreateSocketEvent("user_joined", string(updatedUsers), userID)
		m := message{joinedEvent, ss.arena}
		h.broadcast <- m

		go ss.writePump()
		go ss.readPump(s)
	}
}
