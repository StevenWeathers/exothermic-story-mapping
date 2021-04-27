package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"image"
	"image/png"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/anthonynsimon/bild/transform"
	"github.com/gorilla/mux"
	"github.com/ipsn/go-adorable"
	"github.com/o1egl/govatar"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"
)

type contextKey string

var (
	contextKeyUserID contextKey
	apiKeyHeaderName string = "X-API-Key"
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
		Path:     s.config.PathPrefix + "/",
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
		Path:   s.config.PathPrefix + "/",
		MaxAge: -1,
	}
	beCookie := &http.Cookie{
		Name:     s.config.SecureCookieName,
		Value:    "",
		Path:     s.config.PathPrefix + "/",
		Domain:   s.config.AppDomain,
		Secure:   s.config.SecureCookieFlag,
		SameSite: http.SameSiteStrictMode,
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
		apiKey := r.Header.Get(apiKeyHeaderName)
		apiKey = strings.TrimSpace(apiKey)
		var userID string

		if apiKey != "" {
			var apiKeyErr error
			userID, apiKeyErr = s.database.ValidateAPIKey(apiKey)
			if apiKeyErr != nil {
				log.Println("error validating api key : " + apiKeyErr.Error() + "\n")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		} else {
			var cookieErr error
			userID, cookieErr = s.validateUserCookie(w, r)
			if cookieErr != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		adminErr := s.database.ConfirmAdmin(userID)
		if adminErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)

		h(w, r.WithContext(ctx))
	}
}

// userOnly validates that the request was made by a valid user
func (s *server) userOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get(apiKeyHeaderName)
		apiKey = strings.TrimSpace(apiKey)
		var userID string

		if apiKey != "" {
			var apiKeyErr error
			userID, apiKeyErr = s.database.ValidateAPIKey(apiKey)
			if apiKeyErr != nil {
				log.Println("error validating api key : " + apiKeyErr.Error() + "\n")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		} else {
			var cookieErr error
			userID, cookieErr = s.validateUserCookie(w, r)
			if cookieErr != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		_, warErr := s.database.GetUser(userID)
		if warErr != nil {
			log.Println("error finding user : " + warErr.Error() + "\n")
			s.clearUserCookies(w)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)

		h(w, r.WithContext(ctx))
	}
}

/*
	Handlers
*/

// handleIndex parses the index html file, injecting any relevant data
func (s *server) handleIndex() http.HandlerFunc {
	type AppConfig struct {
		AvatarService             string
		ToastTimeout              int
		AllowGuests               bool
		AllowRegistration         bool
		DefaultLocale             string
		AuthMethod                string
		AppVersion                string
		CookieName                string
		PathPrefix                string
		APIEnabled                bool
		CleanupGuestsDaysOld      int
		CleanupStoryboardsDaysOld int
	}
	type UIConfig struct {
		AnalyticsEnabled bool
		AnalyticsID      string
		AppConfig        AppConfig
	}

	// get the html template from dist, have it ready for requests
	tmplContent, ioErr := fs.ReadFile(f, "dist/index.html")
	if ioErr != nil {
		log.Println("Error opening index template")
		log.Fatal(ioErr)
	}

	tmplString := string(tmplContent)
	tmpl, tmplErr := template.New("index").Parse(tmplString)
	if tmplErr != nil {
		log.Println("Error parsing index template")
		log.Fatal(tmplErr)
	}

	appConfig := AppConfig{
		AvatarService:             viper.GetString("config.avatar_service"),
		ToastTimeout:              viper.GetInt("config.toast_timeout"),
		AllowGuests:               viper.GetBool("config.allow_guests"),
		AllowRegistration:         viper.GetBool("config.allow_registration") && viper.GetString("auth.method") == "normal",
		DefaultLocale:             viper.GetString("config.default_locale"),
		AuthMethod:                viper.GetString("auth.method"),
		APIEnabled:                viper.GetBool("config.allow_external_api"),
		AppVersion:                s.config.Version,
		CookieName:                s.config.FrontendCookieName,
		PathPrefix:                s.config.PathPrefix,
		CleanupGuestsDaysOld:      viper.GetInt("config.cleanup_guests_days_old"),
		CleanupStoryboardsDaysOld: viper.GetInt("config.cleanup_storyboards_days_old"),
	}

	data := UIConfig{
		AnalyticsEnabled: s.config.AnalyticsEnabled,
		AnalyticsID:      s.config.AnalyticsID,
		AppConfig:        appConfig,
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

		authedUser, err := s.authUserDatabase(UserEmail, UserPassword)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		cookie := s.createCookie(authedUser.UserID)
		if cookie != nil {
			http.SetCookie(w, cookie)
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RespondWithJSON(w, http.StatusOK, authedUser)
	}
}

// handleLdapLogin attempts to authenticate the user by looking up and authenticating
// via ldap, and then creates the user if not existing and logs them in
func (s *server) handleLdapLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		keyVal := make(map[string]string)
		json.Unmarshal(body, &keyVal)
		UserEmail := keyVal["userEmail"]
		UserPassword := keyVal["userPassword"]

		authedUser, err := s.authAndCreateUserLdap(UserEmail, UserPassword)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		cookie := s.createCookie(authedUser.UserID)
		if cookie != nil {
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

// handleUserRecruit registers a user as a guest user
func (s *server) handleUserRecruit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		AllowGuests := viper.GetBool("config.allow_guests")
		if !AllowGuests {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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
		AllowRegistration := viper.GetBool("config.allow_registration")
		if !AllowRegistration {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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

// handleForgotPassword attempts to send a password reset email
func (s *server) handleForgotPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors

		keyVal := make(map[string]string)
		json.Unmarshal(body, &keyVal) // check for errors
		UserEmail := keyVal["userEmail"]

		ResetID, UserName, resetErr := s.database.UserResetRequest(UserEmail)
		if resetErr == nil {
			s.email.SendForgotPassword(UserName, UserEmail, ResetID)
		}

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

		userID := r.Context().Value(contextKeyUserID).(string)

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

		userCookieID := r.Context().Value(contextKeyUserID).(string)
		if UserID != userCookieID {
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
		keyVal := make(map[string]interface{})
		json.Unmarshal(body, &keyVal) // check for errors
		UserName := keyVal["userName"].(string)
		UserAvatar := keyVal["userAvatar"].(string)

		UserID := vars["id"]
		userCookieID := r.Context().Value(contextKeyUserID).(string)
		if UserID != userCookieID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		updateErr := s.database.UpdateUserProfile(UserID, UserName, UserAvatar)
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

// handleUserDelete attempts to delete a users account
func (s *server) handleUserDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		UserID := vars["id"]
		userookieID := r.Context().Value(contextKeyUserID).(string)
		if UserID != userookieID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		updateErr := s.database.DeleteUser(UserID)
		if updateErr != nil {
			log.Println("error attempting to delete user : " + updateErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.clearUserCookies(w)

		return
	}
}

// handleUserAvatar creates an avatar for the given user by ID
func (s *server) handleUserAvatar() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		Width, _ := strconv.Atoi(vars["width"])
		UserID := vars["id"]
		AvatarGender := govatar.MALE
		userGender, ok := vars["avatar"]
		if ok {
			if userGender == "female" {
				AvatarGender = govatar.FEMALE
			}
		}

		var avatar image.Image
		if s.config.AvatarService == "govatar" {
			avatar, _ = govatar.GenerateForUsername(AvatarGender, UserID)
		} else { // must be goadorable
			var err error
			avatar, _, err = image.Decode(bytes.NewReader(adorable.PseudoRandom([]byte(UserID))))
			if err != nil {
				log.Fatalln(err)
			}
		}

		img := transform.Resize(avatar, Width, Width, transform.Linear)
		buffer := new(bytes.Buffer)

		if err := png.Encode(buffer, img); err != nil {
			log.Println("unable to encode image.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

		if _, err := w.Write(buffer.Bytes()); err != nil {
			log.Println("unable to write image.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

/*
	Storyboard Handlers
*/

// handleStoryboardCreate handles creating a storyboard (arena)
func (s *server) handleStoryboardCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(contextKeyUserID).(string)

		body, bodyErr := ioutil.ReadAll(r.Body) // check for errors
		if bodyErr != nil {
			log.Println("error in reading request body: " + bodyErr.Error() + "\n")
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
		userID := r.Context().Value(contextKeyUserID).(string)

		storyboards, err := s.database.GetStoryboardsByUser(userID)

		if err != nil {
			http.NotFound(w, r)
			return
		}

		RespondWithJSON(w, http.StatusOK, storyboards)
	}
}

/*
	API Key Handlers
*/

// handleAPIKeyGenerate handles generating an API key for a user
func (s *server) handleAPIKeyGenerate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		body, _ := ioutil.ReadAll(r.Body) // check for errors
		keyVal := make(map[string]interface{})
		json.Unmarshal(body, &keyVal) // check for errors
		APIKeyName := keyVal["name"].(string)

		UserID := vars["id"]
		userCookieID := r.Context().Value(contextKeyUserID).(string)
		if UserID != userCookieID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		APIKey, keyErr := s.database.GenerateAPIKey(UserID, APIKeyName)
		if keyErr != nil {
			log.Println("error attempting to generate api key : " + keyErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RespondWithJSON(w, http.StatusOK, APIKey)
	}
}

// handleUserAPIKeys handles getting user API keys
func (s *server) handleUserAPIKeys() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		UserID := vars["id"]
		userCookieID := r.Context().Value(contextKeyUserID).(string)
		if UserID != userCookieID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		APIKeys, keysErr := s.database.GetUserAPIKeys(UserID)
		if keysErr != nil {
			log.Println("error retrieving api keys : " + keysErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RespondWithJSON(w, http.StatusOK, APIKeys)
	}
}

// handleUserAPIKeys handles getting user API keys
func (s *server) handleUserAPIKeyUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		UserID := vars["id"]
		userCookieID := r.Context().Value(contextKeyUserID).(string)
		if UserID != userCookieID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		APK := vars["keyID"]
		body, _ := ioutil.ReadAll(r.Body) // check for errors
		keyVal := make(map[string]interface{})
		json.Unmarshal(body, &keyVal) // check for errors
		active := keyVal["active"].(bool)

		APIKeys, keysErr := s.database.UpdateUserAPIKey(UserID, APK, active)
		if keysErr != nil {
			log.Println("error updating api key : " + keysErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RespondWithJSON(w, http.StatusOK, APIKeys)
	}
}

// handleUserAPIKeys handles getting user API keys
func (s *server) handleUserAPIKeyDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		UserID := vars["id"]
		userCookieID := r.Context().Value(contextKeyUserID).(string)
		if UserID != userCookieID {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		APK := vars["keyID"]

		APIKeys, keysErr := s.database.DeleteUserAPIKey(UserID, APK)
		if keysErr != nil {
			log.Println("error deleting api key : " + keysErr.Error() + "\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		RespondWithJSON(w, http.StatusOK, APIKeys)
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

// handleGetRegisteredUsers gets a list of registered users
func (s *server) handleGetRegisteredUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		Limit, _ := strconv.Atoi(vars["limit"])
		Offset, _ := strconv.Atoi(vars["offset"])

		Users := s.database.GetRegisteredUsers(Limit, Offset)

		RespondWithJSON(w, http.StatusOK, Users)
	}
}

// handleUserCreate registers a user as a registered user
func (s *server) handleUserCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors
		keyVal := make(map[string]string)
		jsonErr := json.Unmarshal(body, &keyVal) // check for errors
		if jsonErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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

		newUser, VerifyID, err := s.database.CreateUserRegistered(UserName, UserEmail, UserPassword, "")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.email.SendWelcome(UserName, UserEmail, VerifyID)

		RespondWithJSON(w, http.StatusOK, newUser)
	}
}

// handleUserPromote handles promoting a user to ADMIN by ID
func (s *server) handleUserPromote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors
		keyVal := make(map[string]string)
		jsonErr := json.Unmarshal(body, &keyVal) // check for errors
		if jsonErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := s.database.PromoteUser(keyVal["userId"])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}

// handleUserDemote handles demoting a user to REGISTERED by ID
func (s *server) handleUserDemote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body) // check for errors
		keyVal := make(map[string]string)
		jsonErr := json.Unmarshal(body, &keyVal) // check for errors
		if jsonErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := s.database.DemoteUser(keyVal["userId"])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}

// handleCleanStoryboards handles cleaning up old storyboards (ADMIN Manaually Triggered)
func (s *server) handleCleanStoryboards() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		DaysOld := viper.GetInt("config.cleanup_storyboards_days_old")

		err := s.database.CleanStoryboards(DaysOld)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}

// handleCleanGuests handles cleaning up old guests (ADMIN Manaually Triggered)
func (s *server) handleCleanGuests() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		DaysOld := viper.GetInt("config.cleanup_guests_days_old")

		err := s.database.CleanGuests(DaysOld)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
}
