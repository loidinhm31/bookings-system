package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/loidinhm31/access-system/internal/config"
	"github.com/loidinhm31/access-system/internal/constants"
	"github.com/loidinhm31/access-system/internal/driver"
	"github.com/loidinhm31/access-system/internal/handlers"
	"github.com/loidinhm31/access-system/internal/helpers"
	"github.com/loidinhm31/access-system/internal/models"
	"github.com/loidinhm31/access-system/internal/render"
	"log"
	"net/http"
	"os"
	"time"
)

const portNumber = ":8080"

var app config.AppConfig
var sessionManager *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer func(SQL *sql.DB) {
		err := SQL.Close()
		if err != nil {
			log.Fatal("Cannot close database connection", err)
		}
	}(db.SQL)

	log.Println(fmt.Sprintf("Starting application on port %s", portNumber))

	server := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = server.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// Values using in the session
	gob.Register(models.Reservation{})

	// production value
	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = app.InProduction

	app.SessionManager = sessionManager

	// connect to database
	log.Println("Connecting to database...")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=postgrespw")
	if err != nil {
		log.Fatal("Cannot connect to database! Stopping...", err)
	}
	log.Println("Connected to database")

	templateCache, err := render.CreateTemplateCache(constants.PathToTemplate)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	app.TemplateCache = templateCache
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	render.NewTemplates(&app)
	helpers.NewHelpers(&app)

	return db, err
}
