package main

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"net/http"
	"strconv"

	"github.com/anthonynsimon/bild/transform"
	"github.com/gorilla/mux"
	"github.com/ipsn/go-adorable"
	"github.com/o1egl/govatar"
)

// handleUpdatePassword attempts to update a users password
func (s *server) handleUpdatePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keyVal := s.getJSONRequestBody(r, w)

		userID := r.Context().Value(contextKeyUserID).(string)

		UserPassword, passwordErr := ValidateUserPassword(
			keyVal["userPassword1"].(string),
			keyVal["userPassword2"].(string),
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

		s.respondWithJSON(w, http.StatusOK, user)
	}
}

// handleUserProfileUpdate attempts to update users profile (currently limited to name)
func (s *server) handleUserProfileUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		keyVal := s.getJSONRequestBody(r, w)
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
		keyVal := s.getJSONRequestBody(r, w)
		VerifyID := keyVal["verifyId"].(string)

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
