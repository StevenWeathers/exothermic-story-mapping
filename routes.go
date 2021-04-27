package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/spf13/viper"
)

//go:embed dist
var f embed.FS

func (s *server) routes() {
	fsys, err := fs.Sub(f, "dist")
	if err != nil {
		panic(err)
	}
	staticHandler := http.FileServer(http.FS(fsys))

	// static assets
	s.router.PathPrefix("/static/").Handler(http.StripPrefix(s.config.PathPrefix, staticHandler))
	s.router.PathPrefix("/img/").Handler(http.StripPrefix(s.config.PathPrefix, staticHandler))
	s.router.PathPrefix("/lang/").Handler(http.StripPrefix(s.config.PathPrefix, staticHandler))
	// user avatar generation
	if s.config.AvatarService == "goadorable" || s.config.AvatarService == "govatar" {
		s.router.PathPrefix("/avatar/{width}/{id}/{avatar}").Handler(s.handleUserAvatar()).Methods("GET")
		s.router.PathPrefix("/avatar/{width}/{id}").Handler(s.handleUserAvatar()).Methods("GET")
	}
	// api
	// user authentication, profile
	if viper.GetString("auth.method") == "ldap" {
		s.router.HandleFunc("/api/auth", s.handleLdapLogin()).Methods("POST")
	} else {
		s.router.HandleFunc("/api/auth", s.handleLogin()).Methods("POST")
		s.router.HandleFunc("/api/auth/forgot-password", s.handleForgotPassword()).Methods("POST")
		s.router.HandleFunc("/api/auth/reset-password", s.handleResetPassword()).Methods("POST")
		s.router.HandleFunc("/api/auth/update-password", s.userOnly(s.handleUpdatePassword())).Methods("POST")
		s.router.HandleFunc("/api/auth/verify", s.handleAccountVerification()).Methods("POST")
		s.router.HandleFunc("/api/register", s.handleUserEnlist()).Methods("POST")
	}
	s.router.HandleFunc("/api/user", s.handleUserRecruit()).Methods("POST")
	s.router.HandleFunc("/api/auth/logout", s.handleLogout()).Methods("POST")
	s.router.HandleFunc("/api/user/{id}/apikey/{keyID}", s.userOnly(s.handleUserAPIKeyUpdate())).Methods("PUT")
	s.router.HandleFunc("/api/user/{id}/apikey/{keyID}", s.userOnly(s.handleUserAPIKeyDelete())).Methods("DELETE")
	s.router.HandleFunc("/api/user/{id}/apikey", s.userOnly(s.handleAPIKeyGenerate())).Methods("POST")
	s.router.HandleFunc("/api/user/{id}/apikeys", s.userOnly(s.handleUserAPIKeys())).Methods("GET")
	s.router.HandleFunc("/api/user/{id}", s.userOnly(s.handleUserProfile())).Methods("GET")
	s.router.HandleFunc("/api/user/{id}", s.userOnly(s.handleUserProfileUpdate())).Methods("POST")
	s.router.HandleFunc("/api/user/{id}", s.userOnly(s.handleUserDelete())).Methods("DELETE")
	// storyboard(s)
	s.router.HandleFunc("/api/storyboard/{id}", s.handleStoryboardGet())
	s.router.HandleFunc("/api/storyboard", s.userOnly(s.handleStoryboardCreate())).Methods("POST")
	s.router.HandleFunc("/api/storyboards", s.userOnly(s.handleStoryboardsGet()))
	// admin routes
	s.router.HandleFunc("/api/admin/stats", s.adminOnly(s.handleAppStats()))
	s.router.HandleFunc("/api/admin/users/{limit}/{offset}", s.adminOnly(s.handleGetRegisteredUsers()))
	s.router.HandleFunc("/api/admin/user", s.adminOnly(s.handleUserCreate())).Methods("POST")
	s.router.HandleFunc("/api/admin/promote", s.adminOnly(s.handleUserPromote())).Methods("POST")
	s.router.HandleFunc("/api/admin/demote", s.adminOnly(s.handleUserDemote())).Methods("POST")
	s.router.HandleFunc("/api/admin/clean-storyboards", s.adminOnly(s.handleCleanStoryboards())).Methods("DELETE")
	s.router.HandleFunc("/api/admin/clean-guests", s.adminOnly(s.handleCleanGuests())).Methods("DELETE")
	// websocket for storyboard
	s.router.HandleFunc("/api/arena/{id}", s.serveWs())
	// handle index.html
	s.router.PathPrefix("/").HandlerFunc(s.handleIndex())
}
