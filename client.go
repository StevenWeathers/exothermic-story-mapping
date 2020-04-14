package main

import (
	"encoding/json"
	"errors"
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
func (s subscription) readPump() {
	c := s.conn
	defer func() {
		StoryboardID := s.arena
		UserID := s.userID

		Users := RetreatUser(StoryboardID, UserID)
		updatedUsers, _ := json.Marshal(Users)

		retreatEvent := CreateSocketEvent("user_retreated", string(updatedUsers), UserID)
		m := message{retreatEvent, StoryboardID}
		h.broadcast <- m

		h.unregister <- s
		c.ws.Close()
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
			goals, err := CreateStoryboardGoal(storyboardID, userID, keyVal["value"])
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

			goals, err := ReviseGoalName(storyboardID, userID, GoalID, GoalName)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("goal_revised", string(updatedGoals), "")
		case "delete_goal":
			goals, err := DeleteStoryboardGoal(storyboardID, userID, keyVal["value"])
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

			goals, err := CreateStoryboardColumn(storyboardID, GoalID, userID)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("column_added", string(updatedGoals), "")
		case "delete_column":
			goals, err := DeleteStoryboardColumn(storyboardID, userID, keyVal["value"])
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

			goals, err := CreateStoryboardStory(storyboardID, GoalID, ColumnID, userID)
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

			goals, err := ReviseStoryName(storyboardID, userID, StoryID, StoryName)
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

			goals, err := ReviseStoryContent(storyboardID, userID, StoryID, StoryContent)
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

			goals, err := ReviseStoryColor(storyboardID, userID, StoryID, StoryColor)
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

			goals, err := MoveStoryboardStory(storyboardID, userID, StoryID, GoalID, ColumnID, PlaceBefore)
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_moved", string(updatedGoals), "")
		case "delete_story":
			goals, err := DeleteStoryboardStory(storyboardID, userID, keyVal["value"])
			if err != nil {
				badEvent = true
				break
			}
			updatedGoals, _ := json.Marshal(goals)
			msg = CreateSocketEvent("story_deleted", string(updatedGoals), "")
		case "promote_owner":
			storyboard, err := SetStoryboardOwner(storyboardID, userID, keyVal["value"])
			if err != nil {
				badEvent = true
				break
			}

			updatedStoryboard, _ := json.Marshal(storyboard)
			msg = CreateSocketEvent("storyboard_updated", string(updatedStoryboard), "")
		case "concede_storyboard":
			err := DeleteStoryboard(storyboardID, userID)
			if err != nil {
				badEvent = true
				break
			}
			msg = CreateSocketEvent("storyboard_conceded", "", "")
		default:
		}

		if badEvent != true {
			m := message{msg, s.arena}
			h.broadcast <- m
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

// ClearUserCookies wipes the frontend and backend cookies
// used in the event of bad cookie reads
func ClearUserCookies(w http.ResponseWriter) {
	feCookie := &http.Cookie{
		Name:   "user",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	beCookie := &http.Cookie{
		Name:     SecureCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(w, feCookie)
	http.SetCookie(w, beCookie)
}

// ValidateUserCookie returns the userID from secure cookies or errors if failures getting it
func ValidateUserCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	var userID string

	if cookie, err := r.Cookie(SecureCookieName); err == nil {
		var value string
		if err = Sc.Decode(SecureCookieName, cookie.Value, &value); err == nil {
			userID = value
		} else {
			log.Println("error in reading user cookie : " + err.Error() + "\n")
			ClearUserCookies(w)
			return "", errors.New("invalid user cookies")
		}
	} else {
		log.Println("error in reading user cookie : " + err.Error() + "\n")
		ClearUserCookies(w)
		return "", errors.New("invalid user cookies")
	}

	return userID, nil
}

// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyboardID := vars["id"]

	// make sure storyboard is legit
	b, storyboardErr := GetStoryboard(storyboardID)
	if storyboardErr != nil {
		http.NotFound(w, r)
		return
	}
	storyboard, _ := json.Marshal(b)

	// make sure user cookies are valid
	userID, cookieErr := ValidateUserCookie(w, r)
	if cookieErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// make sure user exists
	_, warErr := GetStoryboardUser(storyboardID, userID)

	if warErr != nil {
		log.Println("error finding user : " + warErr.Error() + "\n")
		ClearUserCookies(w)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// upgrade to WebSocket connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	c := &connection{send: make(chan []byte, 256), ws: ws}
	s := subscription{c, storyboardID, userID}
	h.register <- s

	Users, _ := AddUserToStoryboard(s.arena, userID)
	updatedUsers, _ := json.Marshal(Users)

	initEvent := CreateSocketEvent("init", string(storyboard), userID)
	_ = c.write(websocket.TextMessage, initEvent)

	joinedEvent := CreateSocketEvent("user_joined", string(updatedUsers), userID)
	m := message{joinedEvent, s.arena}
	h.broadcast <- m

	go s.writePump()
	s.readPump()
}
