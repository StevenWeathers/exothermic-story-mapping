package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/StevenWeathers/exothermic-story-mapping/pkg/database"
	ldap "github.com/go-ldap/ldap/v3"
	"github.com/spf13/viper"
)

func (s *server) createCookie(userID string) *http.Cookie {
	encoded, err := s.cookie.Encode(s.config.SecureCookieName, userID)
	var NewCookie *http.Cookie

	if err == nil {
		NewCookie = &http.Cookie{
			Name:     s.config.SecureCookieName,
			Value:    encoded,
			Path:     s.config.PathPrefix + "/",
			HttpOnly: true,
			Domain:   s.config.AppDomain,
			MaxAge:   86400 * 30, // 30 days
			Secure:   s.config.SecureCookieFlag,
			SameSite: http.SameSiteStrictMode,
		}
	}
	return NewCookie
}

func (s *server) authUserDatabase(userEmail string, userPassword string) (*database.User, error) {
	authedUser, err := s.database.AuthUser(userEmail, userPassword)
	if err != nil {
		log.Println("Failed authenticating user", userEmail)
	} else if authedUser == nil {
		log.Println("Unknown user", userEmail)
	}
	return authedUser, err
}

// Authenticate using LDAP and if user does not exist, automatically add warror as a verified user
func (s *server) authAndCreateUserLdap(userUsername string, userPassword string) (*database.User, error) {
	var authedUser *database.User
	l, err := ldap.DialURL(viper.GetString("auth.ldap.url"))
	if err != nil {
		log.Println("Failed connecting to ldap server at", viper.GetString("auth.ldap.url"))
		return authedUser, err
	}
	defer l.Close()
	if viper.GetBool("auth.ldap.use_tls") {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			log.Println("Failed securing ldap connection", err)
			return authedUser, err
		}
	}

	if viper.GetString("auth.ldap.bindname") != "" {
		err = l.Bind(viper.GetString("auth.ldap.bindname"), viper.GetString("auth.ldap.bindpass"))
		if err != nil {
			log.Println("Failed binding for authentication:", err)
			return authedUser, err
		}
	}

	searchRequest := ldap.NewSearchRequest(viper.GetString("auth.ldap.basedn"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(viper.GetString("auth.ldap.filter"), userUsername),
		[]string{"dn", viper.GetString("auth.ldap.mail_attr"), viper.GetString("auth.ldap.cn_attr")},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Println("Failed performing ldap search query for", userUsername, ":", err)
		return authedUser, err
	}

	if len(sr.Entries) != 1 {
		log.Println("User", userUsername, "does not exist or too many entries returned")
		return authedUser, errors.New("user not found")
	}

	userdn := sr.Entries[0].DN
	useremail := sr.Entries[0].GetAttributeValue(viper.GetString("auth.ldap.mail_attr"))
	usercn := sr.Entries[0].GetAttributeValue(viper.GetString("auth.ldap.cn_attr"))

	err = l.Bind(userdn, userPassword)
	if err != nil {
		log.Println("Failed authenticating user ", userUsername)
		return authedUser, err
	}

	authedUser, err = s.database.GetUserByEmail(useremail)
	if authedUser == nil {
		log.Println("User", useremail, "does not exist in database, auto-recruit")
		newUser, verifyID, err := s.database.CreateUserRegistered(usercn, useremail, "", "")
		if err != nil {
			log.Println("Failed auto-creating new user", err)
			return authedUser, err
		}
		err = s.database.VerifyUserAccount(verifyID)
		if err != nil {
			log.Println("Failed verifying new user", err)
			return authedUser, err
		}
		authedUser = newUser
	}

	return authedUser, nil
}
