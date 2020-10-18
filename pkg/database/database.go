package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq" // necessary for postgres
	"github.com/markbates/pkger"
	"github.com/spf13/viper"
)

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

// New runs db migrations, sets up a db connection pool
// and sets previously active users to false during startup
func New(AdminEmail string) *Database {
	var d = &Database{
		// read environment variables and sets up mailserver configuration values
		config: &Config{
			host:     viper.GetString("db.host"),
			port:     viper.GetInt("db.port"),
			user:     viper.GetString("db.user"),
			password: viper.GetString("db.pass"),
			dbname:   viper.GetString("db.name"),
			sslmode:  viper.GetString("db.sslmode"),
		},
	}

	sqlFile, ioErr := pkger.Open("/schema.sql")
	if ioErr != nil {
		log.Println("Error reading schema.sql file required to migrate db")
		log.Fatal(ioErr)
	}
	sqlContent, ioErr := ioutil.ReadAll(sqlFile)
	if ioErr != nil {
		// this will hopefully only possibly panic during development as the file is already in memory otherwise
		log.Println("Error reading schema.sql file required to migrate db")
		log.Fatal(ioErr)
	}
	migrationSQL := string(sqlContent)

	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.config.host,
		d.config.port,
		d.config.user,
		d.config.password,
		d.config.dbname,
		d.config.sslmode,
	)

	pdb, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	d.db = pdb

	if _, err := d.db.Exec(migrationSQL); err != nil {
		log.Fatal(err)
	}

	// on server start reset all users to active false for storyboards
	if _, err := d.db.Exec(
		`call deactivate_all_users();`); err != nil {
		log.Println(err)
	}

	// on server start if admin email is specified set that user to ADMIN type
	if AdminEmail != "" {
		if _, err := d.db.Exec(
			`call promote_user_by_email($1);`,
			AdminEmail,
		); err != nil {
			log.Println(err)
		}
	}

	return d
}
