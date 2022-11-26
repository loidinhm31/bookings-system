package main

import (
	"database/sql"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
	"github.com/loidinhm31/go-bookings-system/internal/config"
	"github.com/loidinhm31/go-bookings-system/internal/driver"
	"github.com/loidinhm31/go-bookings-system/internal/handlers"
	"github.com/loidinhm31/go-bookings-system/internal/helpers"
	"github.com/loidinhm31/go-bookings-system/internal/models"
	"github.com/loidinhm31/go-bookings-system/internal/render"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
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

	// get environment
	env := flag.String("env", "dev", "Environment")
	flag.Parse()

	log.Printf("Started with %s profile\n", *env)
	var err error
	if *env == "dev" {
		err = godotenv.Load("./application_local.env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	} else if *env == "stage" {
		err = godotenv.Load("./application_stage.env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	} else if *env == "prod" {
		err = godotenv.Load("./application_prod.env")

		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	// Load env variables
	prodMode := os.Getenv("PROD_MODE")
	productionMode, _ := strconv.ParseBool(prodMode)
	cache := os.Getenv("USE_CACHE")
	useCache, _ := strconv.ParseBool(cache)

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbSsl := os.Getenv("DB_SSL")
	dbName := os.Getenv("DB_NAME")

	// production value
	app.InProduction = productionMode

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
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		dbHost, dbPort, dbName, dbUser, dbPass, dbSsl)
	db, err := driver.ConnectSQL(connStr)
	if err != nil {
		log.Fatal("Cannot connect to database! Stopping...", err)
	}
	log.Println("Connected to database")

	app.PathToTemplate = "./templates"
	app.TemplateCache = map[string]*template.Template{}
	app.UseCache = useCache

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, err
}
