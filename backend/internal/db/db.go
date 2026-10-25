package db

import (
	"database/sql"
	"time"

	// Register the MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

// Open connects to MySQL, configures the connection pool and verifies the
// connection with a ping.
func Open(dsn string) (*sql.DB, error) {
	database, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(20)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(time.Hour)

	if err := database.Ping(); err != nil {
		database.Close()
		return nil, err
	}
	return database, nil
}
