package dbrepo

import (
	"database/sql"
	"github.com/loidinhm31/bookings-system/internal/config"
	"github.com/loidinhm31/bookings-system/internal/repository"
)

type postgresDbRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(conn *sql.DB, app *config.AppConfig) repository.DatabaseRepo {
	return &postgresDbRepo{
		App: app,
		DB:  conn,
	}
}

func NewTestingRepo(testApp *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: testApp,
	}
}
