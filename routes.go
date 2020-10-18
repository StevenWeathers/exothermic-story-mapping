package main

import (
	"net/http"

	"github.com/markbates/pkger"
	"github.com/spf13/viper"
)

func (s *server) routes() {
	staticHandler := http.FileServer(pkger.Dir("/dist"))
	// static assets
	s.router.PathPrefix("/static/").Handler(http.StripPrefix(s.config.PathPrefix, staticHandler))
	s.router.PathPrefix("/img/").Handler(http.StripPrefix(s.config.PathPrefix, staticHandler))
	s.router.PathPrefix("/lang/").Handler(http.StripPrefix(s.config.PathPrefix, staticHandler))
	// user avatar generation
	if s.config.AvatarService == "goadorable" || s.config.AvatarService == "govatar" {
		s.router.PathPrefix("/avatar/{width}/{id}/{avatar}").Handler(s.handleUserAvatar()).Methods("GET")
		s.router.PathPrefix("/avatar/{width}/{id}").Handler(s.handleUserAvatar()).Methods("GET")
	}
	// api (currently internal to UI application)
	// user authentication, profile
	if viper.GetString("auth.method") == "ldap" {
		s.router.HandleFunc("/api/auth", s.handleLdapLogin()).Methods("POST")
	} else {
		s.router.HandleFunc("/api/auth", s.handleLogin()).Methods("POST")
		s.router.HandleFunc("/api/auth/forgot-password", s.handleForgotPassword()).Methods("POST")
		s.router.HandleFunc("/api/auth/reset-password", s.handleResetPassword()).Methods("POST")
		s.router.HandleFunc("/api/auth/update-password", s.handleUpdatePassword()).Methods("POST")
		s.router.HandleFunc("/api/auth/verify", s.handleAccountVerification()).Methods("POST")
		s.router.HandleFunc("/api/register", s.handleUserEnlist()).Methods("POST")
	}
	s.router.HandleFunc("/api/user", s.handleUserRecruit()).Methods("POST")
	s.router.HandleFunc("/api/auth/logout", s.handleLogout()).Methods("POST")
	s.router.HandleFunc("/api/user/{id}", s.handleUserProfile()).Methods("GET")
	s.router.HandleFunc("/api/user/{id}", s.handleUserProfileUpdate()).Methods("POST")
	// storyboard(s)
	s.router.HandleFunc("/api/storyboard", s.handleStoryboardCreate()).Methods("POST")
	s.router.HandleFunc("/api/storyboard/{id}", s.handleStoryboardGet())
	s.router.HandleFunc("/api/storyboards", s.handleStoryboardsGet())
	// admin routes
	s.router.HandleFunc("/api/admin/stats", s.adminOnly(s.handleAppStats()))
	s.router.HandleFunc("/api/admin/users", s.adminOnly(s.handleGetRegisteredUsers()))
	s.router.HandleFunc("/api/admin/user", s.adminOnly(s.handleUserCreate())).Methods("POST")
	// websocket for storyboard
	s.router.HandleFunc("/api/arena/{id}", s.serveWs())
	// handle index.html
	s.router.PathPrefix("/").HandlerFunc(s.handleIndex())
}
