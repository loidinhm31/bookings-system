package config

import (
	"github.com/alexedwards/scs/v2"
	"github.com/loidinhm31/bookings-system/internal/models"
	"html/template"
	"log"
)

type AppConfig struct {
	UseCache       bool
	TemplateCache  map[string]*template.Template
	PathToTemplate string
	InfoLog        *log.Logger
	ErrorLog       *log.Logger
	InProduction   bool
	SessionManager *scs.SessionManager
	MailChannel    chan models.MailData
}
