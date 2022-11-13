package main

import (
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/loidinhm31/access-system/internal/config"
	"github.com/loidinhm31/access-system/internal/handlers"
	"github.com/loidinhm31/access-system/internal/models"
	"github.com/loidinhm31/access-system/internal/render"
	"log"
	"net/http"
	"time"
)

const portNumber = ":8080"

var app config.AppConfig
var sessionManager *scs.SessionManager

func main() {
	// Values using in the session
	gob.Register(models.Reservation{})

	// production value
	app.InProduction = false

	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = app.InProduction

	app.SessionManager = sessionManager

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal(err)
	}

	app.TemplateCache = templateCache
	app.UseCache = false

	repo := handlers.CreateNewRepo(&app)
	handlers.SetRepository(repo)

	render.NewTemplates(&app)

	log.Println(fmt.Sprintf("Starting application on port %s", portNumber))

	server := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = server.ListenAndServe()
	log.Fatal(err)
}
