package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB holds the database connection pool
type DB struct {
	SQL *sql.DB
}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifeTime = 5 * time.Minute

var dbConn = &DB{}

// ConnectSQL creates database pool
func ConnectSQL(dsn string) (*DB, error) {
	db, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(maxOpenDbConn)
	db.SetConnMaxIdleTime(maxIdleDbConn)
	db.SetConnMaxLifetime(maxDbLifeTime)

	dbConn.SQL = db

	err = testDb(db)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

// NewDatabase creates a new database for the application
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// try to ping the database
func testDb(db *sql.DB) error {
	err := db.Ping()
	if err != nil {
		return err
	}
	return nil
}
