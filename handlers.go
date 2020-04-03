package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// LoginHandler attempts to login the user by comparing email/password to whats in DB
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body) // check for errors

	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal) // check for errors
	UserEmail := keyVal["userEmail"]
	UserPassword := keyVal["userPassword"]

	authedUser, err := AuthUser(UserEmail, UserPassword)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	encoded, err := Sc.Encode(SecureCookieName, authedUser.UserID)
	if err == nil {
		cookie := &http.Cookie{
			Name:     SecureCookieName,
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
			Domain:   AppDomain,
			MaxAge:   86400 * 30, // 30 days
			Secure:   SecureCookieFlag,
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

// LogoutHandler clears the user cookie(s) ending session
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearUserCookies(w)
	return
}

// CreateStoryboardHandler handles creating a storyboard (arena)
func CreateStoryboardHandler(w http.ResponseWriter, r *http.Request) {
	userID, cookieErr := ValidateUserCookie(w, r)
	if cookieErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, warErr := GetUser(userID)

	if warErr != nil {
		log.Println("error finding user : " + warErr.Error() + "\n")
		ClearUserCookies(w)
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

	newStoryboard, err := CreateStoryboard(userID, keyVal.StoryboardName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	RespondWithJSON(w, http.StatusOK, newStoryboard)
}

// RecruitUserHandler registers a user as a guest user
func RecruitUserHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body) // check for errors

	keyVal := make(map[string]string)
	jsonErr := json.Unmarshal(body, &keyVal) // check for errors
	if jsonErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	UserName := keyVal["userName"]

	newUser, err := CreateUserGuest(UserName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoded, err := Sc.Encode(SecureCookieName, newUser.UserID)
	if err == nil {
		cookie := &http.Cookie{
			Name:     SecureCookieName,
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
			Domain:   AppDomain,
			MaxAge:   86400 * 365, // 365 days
			Secure:   SecureCookieFlag,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)
	} else {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	RespondWithJSON(w, http.StatusOK, newUser)
}

// EnlistUserHandler registers a user as a registered user (authenticated)
func EnlistUserHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body) // check for errors
	keyVal := make(map[string]string)
	jsonErr := json.Unmarshal(body, &keyVal) // check for errors
	if jsonErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ActiveUserID, _ := ValidateUserCookie(w, r)

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

	newUser, VerifyID, err := CreateUserRegistered(UserName, UserEmail, UserPassword, ActiveUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoded, err := Sc.Encode(SecureCookieName, newUser.UserID)
	if err == nil {
		cookie := &http.Cookie{
			Name:     SecureCookieName,
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
			Domain:   AppDomain,
			MaxAge:   86400 * 30, // 30 days
			Secure:   SecureCookieFlag,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)
	} else {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	SendWelcomeEmail(UserName, UserEmail, VerifyID)

	RespondWithJSON(w, http.StatusOK, newUser)
}

// GetStoryboardHandler looks up storyboard or returns notfound status
func GetStoryboardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	StoryboardID := vars["id"]

	storyboard, err := GetStoryboard(StoryboardID)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	RespondWithJSON(w, http.StatusOK, storyboard)
}

// GetStoryboardsHandler looks up storyboards associated with userID
func GetStoryboardsHandler(w http.ResponseWriter, r *http.Request) {
	userID, cookieErr := ValidateUserCookie(w, r)
	if cookieErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, warErr := GetUser(userID)

	if warErr != nil {
		log.Println("error finding user : " + warErr.Error() + "\n")
		ClearUserCookies(w)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	storyboards, err := GetStoryboardsByUser(userID)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	RespondWithJSON(w, http.StatusOK, storyboards)
}

// ForgotPasswordHandler attempts to send a password reset email
func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body) // check for errors

	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal) // check for errors
	UserEmail := keyVal["userEmail"]

	ResetID, UserName, resetErr := UserResetRequest(UserEmail)
	if resetErr != nil {
		log.Println("error attempting to send user reset : " + resetErr.Error() + "\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	SendForgotPasswordEmail(UserName, UserEmail, ResetID)

	w.WriteHeader(http.StatusOK)
	return
}

// ResetPasswordHandler attempts to reset a users password
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
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

	UserName, UserEmail, resetErr := UserResetPassword(ResetID, UserPassword)
	if resetErr != nil {
		log.Println("error attempting to reset user password : " + resetErr.Error() + "\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	SendPasswordResetEmail(UserName, UserEmail)

	return
}

// UpdatePasswordHandler attempts to update a users password
func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body) // check for errors
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal) // check for errors

	userID, cookieErr := ValidateUserCookie(w, r)
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

	UserName, UserEmail, updateErr := UserUpdatePassword(userID, UserPassword)
	if updateErr != nil {
		log.Println("error attempting to update user password : " + updateErr.Error() + "\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	SendPasswordUpdateEmail(UserName, UserEmail)

	return
}

// UserProfileHandler returns the users profile if it matches their session
func UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	UserID := vars["id"]

	userCookieID, cookieErr := ValidateUserCookie(w, r)
	if cookieErr != nil || UserID != userCookieID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, warErr := GetUser(UserID)
	if warErr != nil {
		log.Println("error finding user : " + warErr.Error() + "\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	RespondWithJSON(w, http.StatusOK, user)
}

// UserUpdateProfileHandler attempts to update users profile (currently limited to name)
func UserUpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	body, _ := ioutil.ReadAll(r.Body) // check for errors
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal) // check for errors
	UserName := keyVal["userName"]

	UserID := vars["id"]
	userCookieID, cookieErr := ValidateUserCookie(w, r)
	if cookieErr != nil || UserID != userCookieID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	updateErr := UpdateUserProfile(UserID, UserName)
	if updateErr != nil {
		log.Println("error attempting to update user profile : " + updateErr.Error() + "\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

// GetAppStatsHandler gets the applications stats
func GetAppStatsHandler(w http.ResponseWriter, r *http.Request) {
	userID, cookieErr := ValidateUserCookie(w, r)
	if cookieErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	AppStats, err := GetAppStats(userID)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	RespondWithJSON(w, http.StatusOK, AppStats)
}

// VerifyAccountHandler attempts to verify a users account
func VerifyAccountHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body) // check for errors

	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal) // check for errors
	VerifyID := keyVal["verifyId"]

	verifyErr := VerifyUserAccount(VerifyID)
	if verifyErr != nil {
		log.Println("error attempting to verify user account : " + verifyErr.Error() + "\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}
