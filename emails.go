package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"strconv"
	"time"

	"github.com/matcornic/hermes/v2"
)

// smtpServer data to smtp server
type smtpServer struct {
	host string
	port string
}

// Address URI to smtp server
func (s *smtpServer) Address() string {
	return s.host + ":" + s.port
}

var smtpHost string
var smtpPort string
var smtpSecure bool
var smtpIdentity string
var smtpUser string
var smtpPass string
var smtpSender string
var smtpServerConfig = smtpServer{}
var tlsConfig = &tls.Config{}
var smtpFrom = mail.Address{}
var smtpAuth smtp.Auth

// GetMailserverConfig reads environment variables and sets up mailserver configuration values
func GetMailserverConfig() {
	smtpHost = GetEnv("SMTP_HOST", "localhost")
	smtpPort = GetEnv("SMTP_PORT", "25")
	smtpSecure = GetBoolEnv("SMTP_SECURE", true)
	smtpIdentity = GetEnv("SMTP_IDENTITY", "")
	smtpUser = GetEnv("SMTP_USER", "")
	smtpPass = GetEnv("SMTP_PASS", "")
	smtpSender = GetEnv("SMTP_SENDER", "no-reply@exothermic.dev")

	// smtp server configuration.
	smtpServerConfig = smtpServer{host: smtpHost, port: smtpPort}

	// smtp sender info
	smtpFrom = mail.Address{
		Name:    "Exothermic",
		Address: smtpSender,
	}

	// TLS config
	tlsConfig = &tls.Config{
		InsecureSkipVerify: !smtpSecure,
		ServerName:         smtpHost,
	}

	smtpAuth = smtp.PlainAuth(smtpIdentity, smtpUser, smtpPass, smtpHost)
}

// Generates an Email Body with hermes
func generateEmailBody(Body hermes.Body) (emailBody string, generateErr error) {
	currentTime := time.Now()
	year := strconv.Itoa(currentTime.Year())
	hms := hermes.Hermes{
		Product: hermes.Product{
			Name:      "Exothermic",
			Link:      "https://" + AppDomain + "/",
			Logo:      "https://" + AppDomain + "/img/exothermic-logo.png",
			Copyright: "Copyright © " + year + " Exothermic. All rights reserved.",
		},
	}

	email := hermes.Email{
		Body: Body,
	}

	// Generate an HTML email with the provided contents (for modern clients)
	emailBody, err := hms.GenerateHTML(email)
	if err != nil {
		return "", err
	}

	return emailBody, nil
}

// utility function to send emails
func sendEmail(UserName string, UserEmail string, Subject string, Body string) error {
	to := mail.Address{
		Name:    UserName,
		Address: UserEmail,
	}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = smtpFrom.String()
	headers["To"] = to.String()
	headers["Subject"] = Subject
	headers["MIME-version"] = "1.0"
	headers["Content-Type"] = "text/html"

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + Body

	c, err := smtp.Dial(smtpServerConfig.Address())
	if err != nil {
		log.Println("Error dialing SMTP: ", err)
		return err
	}

	c.StartTLS(tlsConfig)

	// Auth
	if smtpSecure == true {
		if err = c.Auth(smtpAuth); err != nil {
			log.Println("Error authenticating SMTP: ", err)
			return err
		}
	}

	// To && From
	if err = c.Mail(smtpFrom.Address); err != nil {
		log.Println("Error setting SMTP from: ", err)
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		log.Println("Error setting SMTP to: ", err)
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		log.Println("Error setting SMTP data: ", err)
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		log.Println("Error sending email: ", err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Println("Error closing SMTP: ", err)
		return err
	}

	c.Quit()

	return nil
}

// SendWelcomeEmail sends the welcome email to new registered user
func SendWelcomeEmail(UserName string, UserEmail string, VerifyID string) error {
	emailBody, err := generateEmailBody(
		hermes.Body{
			Name: UserName,
			Intros: []string{
				"Welcome to the Exothermic.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Please validate your email, the following link will expire in 24 hours.",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Verify Account",
						Link:  "https://" + AppDomain + "/verify-account/" + VerifyID,
					},
				},
				{
					Instructions: "Need help, or have questions? Visit our Github page",
					Button: hermes.Button{
						Text: "Github Repo",
						Link: "https://github.com/StevenWeathers/exothermic-story-mapping/",
					},
				},
			},
		},
	)
	if err != nil {
		log.Println("Error Generating Welcome Email HTML: ", err)
		return err
	}

	sendErr := sendEmail(
		UserName,
		UserEmail,
		"Welcome to the Exothermic!",
		emailBody,
	)
	if sendErr != nil {
		log.Println("Error sending Welcome Email: ", sendErr)
		return sendErr
	}

	return nil
}

// SendForgotPasswordEmail Sends a Forgot Password reset email to user
func SendForgotPasswordEmail(UserName string, UserEmail string, ResetID string) error {
	emailBody, err := generateEmailBody(
		hermes.Body{
			Name: UserName,
			Intros: []string{
				"It seems you've forgot your Exothermic password.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Reset your password now, the following link will expire within an hour of the original request.",
					Button: hermes.Button{
						Text: "Reset Password",
						Link: "https://" + AppDomain + "/reset-password/" + ResetID,
					},
				},
				{
					Instructions: "Need help, or have questions? Visit our Github page",
					Button: hermes.Button{
						Text: "Github Repo",
						Link: "https://github.com/StevenWeathers/exothermic-story-mapping/",
					},
				},
			},
		},
	)
	if err != nil {
		log.Println("Error Generating Forgot Password Email HTML: ", err)
		return err
	}

	sendErr := sendEmail(
		UserName,
		UserEmail,
		"Forgot your Exothermic password?",
		emailBody,
	)
	if sendErr != nil {
		log.Println("Error sending Forgot Password Email: ", sendErr)
		return sendErr
	}

	return nil
}

// SendPasswordResetEmail Sends a Reset Password confirmation email to user
func SendPasswordResetEmail(UserName string, UserEmail string) error {
	emailBody, err := generateEmailBody(
		hermes.Body{
			Name: UserName,
			Intros: []string{
				"Your Exothermic password was succesfully reset.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Need help, or have questions? Visit our Github page",
					Button: hermes.Button{
						Text: "Github Repo",
						Link: "https://github.com/StevenWeathers/exothermic-story-mapping/",
					},
				},
			},
		},
	)
	if err != nil {
		log.Println("Error Generating Reset Password Email HTML: ", err)
		return err
	}

	sendErr := sendEmail(
		UserName,
		UserEmail,
		"Your Exothermic password was succesfully reset.",
		emailBody,
	)
	if sendErr != nil {
		log.Println("Error sending Reset Password Email: ", sendErr)
		return sendErr
	}

	return nil
}

// SendPasswordUpdateEmail Sends an Update Password confirmation email to user
func SendPasswordUpdateEmail(UserName string, UserEmail string) error {
	emailBody, err := generateEmailBody(
		hermes.Body{
			Name: UserName,
			Intros: []string{
				"Your Exothermic password was succesfully updated.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Need help, or have questions? Visit our Github page",
					Button: hermes.Button{
						Text: "Github Repo",
						Link: "https://github.com/StevenWeathers/exothermic-story-mapping/",
					},
				},
			},
		},
	)
	if err != nil {
		log.Println("Error Generating Update Password Email HTML: ", err)
		return err
	}

	sendErr := sendEmail(
		UserName,
		UserEmail,
		"Your Exothermic password was succesfully updated.",
		emailBody,
	)
	if sendErr != nil {
		log.Println("Error sending Update Password Email: ", sendErr)
		return sendErr
	}

	return nil
}
