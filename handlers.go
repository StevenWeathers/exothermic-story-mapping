package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/markbates/pkger"
	"gopkg.in/go-playground/validator.v9"
)

type userAccount struct {
	Name      string `json:"name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password1 string `json:"password1" validate:"required,min=6,max=72"`
	Password2 string `json:"password2" validate:"required,min=6,max=72,eqfield=Password1"`
}

type userPassword struct {
	Password1 string `json:"password1" validate:"required,min=6,max=72"`
	Password2 string `json:"password2" validate:"required,min=6,max=72,eqfield=Password1"`
}

// ValidateUserAccount makes sure user name, email, and password are valid before creating the account
func ValidateUserAccount(name string, email string, pwd1 string, pwd2 string) (UserName string, UserEmail string, UserPassword string, validateErr error) {
	v := validator.New()
	a := userAccount{
		Name:      name,
		Email:     email,
		Password1: pwd1,
		Password2: pwd2,
	}
	err := v.Struct(a)

	return name, email, pwd1, err
}

// ValidateUserPassword makes sure user password is valid before updating the password
func ValidateUserPassword(pwd1 string, pwd2 string) (UserPassword string, validateErr error) {
	v := validator.New()
	a := userPassword{
		Password1: pwd1,
		Password2: pwd2,
	}
	err := v.Struct(a)

	return pwd1, err
}

// RespondWithJSON takes a payload and writes the response
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// createUserCookie creates the users cookie
func (s *server) createUserCookie(w http.ResponseWriter, isRegistered bool, UserID string) {
	var cookiedays = 365 // 356 days
	if isRegistered == true {
		cookiedays = 30 // 30 days
	}

	encoded, err := s.cookie.Encode(s.config.SecureCookieName, UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return

	}

	cookie := &http.Cookie{
		Name:     s.config.SecureCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Domain:   s.config.AppDomain,
		MaxAge:   86400 * cookiedays,
		Secure:   s.config.SecureCookieFlag,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
}

// clearUserCookies wipes the frontend and backend cookies
// used in the event of bad cookie reads
func (s *server) clearUserCookies(w http.ResponseWriter) {
	feCookie := &http.Cookie{
		Name:   s.config.FrontendCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	beCookie := &http.Cookie{
		Name:     s.config.SecureCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(w, feCookie)
	http.SetCookie(w, beCookie)
}

// validateUserCookie returns the userID from secure cookies or errors if failures getting it
func (s *server) validateUserCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	var userID string

	if cookie, err := r.Cookie(s.config.SecureCookieName); err == nil {
		var value string
		if err = s.cookie.Decode(s.config.SecureCookieName, cookie.Value, &value); err == nil {
			userID = value
		} else {
			log.Println("error in reading user cookie : " + err.Error() + "\n")
			s.clearUserCookies(w)
			return "", errors.New("invalid user cookies")
		}
	} else {
		log.Println("error in reading user cookie : " + err.Error() + "\n")
		s.clearUserCookies(w)
		return "", errors.New("invalid user cookies")
	}

	return userID, nil
}

/*
	Middlewares
*/

// adminOnly middleware checks if the user is an admin, otherwise reject their request
func (s *server) adminOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, cookieErr := s.validateUserCookie(w, r)
		if cookieErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		adminErr := s.database.ConfirmAdmin(userID)
		if adminErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
}

/*
	Handlers
*/

// handleIndex parses the index html file, injecting any relevant data
func (s *server) handleIndex() http.HandlerFunc {
	type UIConfig struct {
		AnalyticsEnabled bool
		AnalyticsID      string
	}

	// get the html template from dist, have it ready for requests
	indexFile, ioErr := pkger.Open("/dist/index.html")
	if ioErr != nil {
		log.Println("Error opening index template")
		log.Fatal(ioErr)
	}
	tmplContent, ioReadErr := ioutil.ReadAll(indexFile)
	if ioReadErr != nil {
		// this will hopefully only possibly panic during development as the file is already in memory otherwise
		log.Println("Error reading index template")
		log.Fatal(ioReadErr)
	}

	tmplString := string(tmplContent)
	tmpl, tmplErr := template.New("index").Parse(tmplString)
	if tmplErr != nil {
		log.Println("Error parsing index template")
		log.Fatal(tmplErr)
	}

	data := UIConfig{
		AnalyticsEnabled: s.config.AnalyticsEnabled,
		AnalyticsID:      s.config.AnalyticsID,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, data)
	}
}

// handleLogin attempts to login the user by comparing email/password to whats in DB
func (s *server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors

		keyVal := make(map[string]string)
		json.Unmarshal(body, &keyVal) // check for errors
		UserEmail := keyVal["userEmail"]
		UserPassword := keyVal["userPassword"]

		authedUser, err := s.database.AuthUser(UserEmail, UserPassword)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		encoded, err := s.cookie.Encode(s.config.SecureCookieName, authedUser.UserID)
		if err == nil {
			cookie := &http.Cookie{
				Name:     s.config.SecureCookieName,
				Value:    encoded,
				Path:     "/",
				HttpOnly: true,
				Domain:   s.config.AppDomain,
				MaxAge:   86400 * 30, // 30 days
				Secure:   s.config.SecureCookieFlag,
				SameSite: http.SameSiteStrictMode,
			}
			http.SetCookie(w, cookie)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RespondWithJSON(w, http.StatusOK, authedUser)
	}
}

// handleLogout clears the user cookie(s) ending session
func (s *server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.clearUserCookies(w)
		return
	}
}

// handleStoryboardCreate handles creating a storyboard (arena)
func (s *server) handleStoryboardCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, cookieErr := s.validateUserCookie(w, r)
		if cookieErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, warErr := s.database.GetUser(userID)

		if warErr != nil {
			log.Println("error finding user : " + warErr.Error() + "\n")
			s.clearUserCookies(w)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		body, bodyErr := ioutil.ReadAll(r.Body) // check for errors
		if bodyErr != nil {
			log.Println("error in reading user cookie : " + bodyErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var keyVal struct {
			StoryboardName string `json:"storyboardName"`
		}
		json.Unmarshal(body, &keyVal) // check for errors

		newStoryboard, err := s.database.CreateStoryboard(userID, keyVal.StoryboardName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RespondWithJSON(w, http.StatusOK, newStoryboard)
	}
}

// handleUserRecruit registers a user as a guest user
func (s *server) handleUserRecruit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors

		keyVal := make(map[string]string)
		jsonErr := json.Unmarshal(body, &keyVal) // check for errors
		if jsonErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		UserName := keyVal["userName"]

		newUser, err := s.database.CreateUserGuest(UserName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.createUserCookie(w, false, newUser.UserID)

		RespondWithJSON(w, http.StatusOK, newUser)
	}
}

// handleUserEnlist registers a user as a registered user (authenticated)
func (s *server) handleUserEnlist() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors
		keyVal := make(map[string]string)
		jsonErr := json.Unmarshal(body, &keyVal) // check for errors
		if jsonErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ActiveUserID, _ := s.validateUserCookie(w, r)

		UserName, UserEmail, UserPassword, accountErr := ValidateUserAccount(
			keyVal["userName"],
			keyVal["userEmail"],
			keyVal["userPassword1"],
			keyVal["userPassword2"],
		)

		if accountErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newUser, VerifyID, err := s.database.CreateUserRegistered(UserName, UserEmail, UserPassword, ActiveUserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.createUserCookie(w, true, newUser.UserID)

		s.email.SendWelcome(UserName, UserEmail, VerifyID)

		RespondWithJSON(w, http.StatusOK, newUser)
	}
}

// handleStoryboardGet looks up storyboard or returns notfound status
func (s *server) handleStoryboardGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		StoryboardID := vars["id"]

		storyboard, err := s.database.GetStoryboard(StoryboardID)

		if err != nil {
			http.NotFound(w, r)
			return
		}

		RespondWithJSON(w, http.StatusOK, storyboard)
	}
}

// handleStoryboardsGet looks up storyboards associated with userID
func (s *server) handleStoryboardsGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, cookieErr := s.validateUserCookie(w, r)
		if cookieErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, warErr := s.database.GetUser(userID)

		if warErr != nil {
			log.Println("error finding user : " + warErr.Error() + "\n")
			s.clearUserCookies(w)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		storyboards, err := s.database.GetStoryboardsByUser(userID)

		if err != nil {
			http.NotFound(w, r)
			return
		}

		RespondWithJSON(w, http.StatusOK, storyboards)
	}
}

// handleForgotPassword attempts to send a password reset email
func (s *server) handleForgotPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors

		keyVal := make(map[string]string)
		json.Unmarshal(body, &keyVal) // check for errors
		UserEmail := keyVal["userEmail"]

		ResetID, UserName, resetErr := s.database.UserResetRequest(UserEmail)
		if resetErr != nil {
			log.Println("error attempting to send user reset : " + resetErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.email.SendForgotPassword(UserName, UserEmail, ResetID)

		w.WriteHeader(http.StatusOK)
		return
	}
}

// handleResetPassword attempts to reset a users password
func (s *server) handleResetPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors

		keyVal := make(map[string]string)
		json.Unmarshal(body, &keyVal) // check for errors
		ResetID := keyVal["resetId"]

		UserPassword, passwordErr := ValidateUserPassword(
			keyVal["userPassword1"],
			keyVal["userPassword2"],
		)

		if passwordErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		UserName, UserEmail, resetErr := s.database.UserResetPassword(ResetID, UserPassword)
		if resetErr != nil {
			log.Println("error attempting to reset user password : " + resetErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.email.SendPasswordReset(UserName, UserEmail)

		return
	}
}

// handleUpdatePassword attempts to update a users password
func (s *server) handleUpdatePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors
		keyVal := make(map[string]string)
		json.Unmarshal(body, &keyVal) // check for errors

		userID, cookieErr := s.validateUserCookie(w, r)
		if cookieErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		UserPassword, passwordErr := ValidateUserPassword(
			keyVal["userPassword1"],
			keyVal["userPassword2"],
		)

		if passwordErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		UserName, UserEmail, updateErr := s.database.UserUpdatePassword(userID, UserPassword)
		if updateErr != nil {
			log.Println("error attempting to update user password : " + updateErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.email.SendPasswordUpdate(UserName, UserEmail)

		return
	}
}

// handleUserProfile returns the users profile if it matches their session
func (s *server) handleUserProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		UserID := vars["id"]

		userCookieID, cookieErr := s.validateUserCookie(w, r)
		if cookieErr != nil || UserID != userCookieID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		user, warErr := s.database.GetUser(UserID)
		if warErr != nil {
			log.Println("error finding user : " + warErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RespondWithJSON(w, http.StatusOK, user)
	}
}

// handleUserProfileUpdate attempts to update users profile (currently limited to name)
func (s *server) handleUserProfileUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		body, _ := ioutil.ReadAll(r.Body) // check for errors
		keyVal := make(map[string]string)
		json.Unmarshal(body, &keyVal) // check for errors
		UserName := keyVal["userName"]

		UserID := vars["id"]
		userCookieID, cookieErr := s.validateUserCookie(w, r)
		if cookieErr != nil || UserID != userCookieID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		updateErr := s.database.UpdateUserProfile(UserID, UserName)
		if updateErr != nil {
			log.Println("error attempting to update user profile : " + updateErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}

// handleAccountVerification attempts to verify a users account
func (s *server) handleAccountVerification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors

		keyVal := make(map[string]string)
		json.Unmarshal(body, &keyVal) // check for errors
		VerifyID := keyVal["verifyId"]

		verifyErr := s.database.VerifyUserAccount(VerifyID)
		if verifyErr != nil {
			log.Println("error attempting to verify user account : " + verifyErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}

/*
	Admin Handlers
*/

// handleAppStats gets the applications stats
func (s *server) handleAppStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		AppStats, err := s.database.GetAppStats()

		if err != nil {
			http.NotFound(w, r)
			return
		}

		RespondWithJSON(w, http.StatusOK, AppStats)
	}
}
