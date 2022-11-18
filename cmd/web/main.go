package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/loidinhm31/bookings-system/internal/config"
	"github.com/loidinhm31/bookings-system/internal/driver"
	"github.com/loidinhm31/bookings-system/internal/handlers"
	"github.com/loidinhm31/bookings-system/internal/helpers"
	"github.com/loidinhm31/bookings-system/internal/models"
	"github.com/loidinhm31/bookings-system/internal/render"
	"html/template"
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

	defer close(app.MailChannel)
	listenForMail()

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
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	mailChannel := make(chan models.MailData)
	app.MailChannel = mailChannel

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
	db, err := driver.ConnectSQL("host=127.0.0.1 port=5432 dbname=bookings user=postgres password=postgrespw")
	if err != nil {
		log.Fatal("Cannot connect to database! Stopping...", err)
	}
	log.Println("Connected to database")

	app.PathToTemplate = "./templates"
	app.TemplateCache = map[string]*template.Template{}
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, err
}
