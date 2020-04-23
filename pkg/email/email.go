package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"os"
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

var smtpServerConfig = smtpServer{}
var tlsConfig = &tls.Config{}
var smtpFrom = mail.Address{}
var smtpAuth smtp.Auth

// Config contains all the mailserver values
type Config struct {
	AppDomain    string
	SenderName   string
	smtpHost     string
	smtpPort     string
	smtpSecure   bool
	smtpIdentity string
	smtpUser     string
	smtpPass     string
	smtpSender   string
}

// Email contains all the methods to send application emails
type Email struct {
	config *Config
}

// GetEnv gets environment variable matching key string
// and if it finds none uses fallback string
// returning either the matching or fallback string
func GetEnv(key string, fallback string) string {
	var result = os.Getenv(key)

	if result == "" {
		result = fallback
	}

	return result
}

// GetIntEnv gets an environment variable and converts it to an int
// and if it finds none uses fallback
func GetIntEnv(key string, fallback int) int {
	var intResult = fallback
	var stringResult = os.Getenv(key)

	if stringResult != "" {
		v, _ := strconv.Atoi(stringResult)
		intResult = v
	}

	return intResult
}

// GetBoolEnv gets an environment variable and converts it to a bool
// and if it finds none uses fallback
func GetBoolEnv(key string, fallback bool) bool {
	var boolResult = fallback
	var stringResult = os.Getenv(key)

	if stringResult != "" {
		b, _ := strconv.ParseBool(stringResult)
		boolResult = b
	}

	return boolResult
}

// New creates a new instance of Email
func New(AppDomain string) *Email {
	var m = &Email{
		// read environment variables and sets up mailserver configuration values
		config: &Config{
			AppDomain:    AppDomain,
			SenderName:   "Exothermic",
			smtpHost:     GetEnv("SMTP_HOST", "localhost"),
			smtpPort:     GetEnv("SMTP_PORT", "25"),
			smtpSecure:   GetBoolEnv("SMTP_SECURE", true),
			smtpIdentity: GetEnv("SMTP_IDENTITY", ""),
			smtpUser:     GetEnv("SMTP_USER", ""),
			smtpPass:     GetEnv("SMTP_PASS", ""),
			smtpSender:   GetEnv("SMTP_SENDER", "no-reply@exothermic.dev"),
		},
	}

	// smtp server configuration.
	smtpServerConfig = smtpServer{host: m.config.smtpHost, port: m.config.smtpPort}

	// smtp sender info
	smtpFrom = mail.Address{
		Name:    m.config.SenderName,
		Address: m.config.smtpSender,
	}

	// TLS config
	tlsConfig = &tls.Config{
		InsecureSkipVerify: !m.config.smtpSecure,
		ServerName:         m.config.smtpHost,
	}

	smtpAuth = smtp.PlainAuth(m.config.smtpIdentity, m.config.smtpUser, m.config.smtpPass, m.config.smtpHost)

	return m
}

// Generates an Email Body with hermes
func (m *Email) generateBody(Body hermes.Body) (emailBody string, generateErr error) {
	currentTime := time.Now()
	year := strconv.Itoa(currentTime.Year())
	hms := hermes.Hermes{
		Product: hermes.Product{
			Name:      "Exothermic",
			Link:      "https://" + m.config.AppDomain + "/",
			Logo:      "https://" + m.config.AppDomain + "/img/exothermic-logo.png",
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

// Send - utility function to send emails
func (m *Email) Send(UserName string, UserEmail string, Subject string, Body string) error {
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
	if m.config.smtpSecure == true {
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
