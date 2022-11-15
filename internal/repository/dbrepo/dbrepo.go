package dbrepo

import (
	"database/sql"
	"github.com/loidinhm31/access-system/internal/config"
	"github.com/loidinhm31/access-system/internal/repository"
)

type postgresDbRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(conn *sql.DB, app *config.AppConfig) repository.DatabaseRepo {
	return &postgresDbRepo{
		App: app,
		DB:  conn,
	}
}
