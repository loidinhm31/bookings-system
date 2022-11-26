package handlers

import (
	"encoding/gob"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
	"github.com/loidinhm31/go-bookings-system/internal/config"
	"github.com/loidinhm31/go-bookings-system/internal/helpers"
	"github.com/loidinhm31/go-bookings-system/internal/models"
	"github.com/loidinhm31/go-bookings-system/internal/render"
	"html/template"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var testApp config.AppConfig
var sessionManager *scs.SessionManager

var functions = template.FuncMap{
	"simpleDate": render.SimpleDate,
	"formatDate": render.FormatDate,
	"iterate":    render.Iterate,
	"add":        render.Add,
}

func TestMain(m *testing.M) {
	/**
	From main.go
	START
	*/
	// Values using in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	// production value
	testApp.InProduction = false

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	testApp.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	testApp.ErrorLog = errorLog

	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = testApp.InProduction

	testApp.SessionManager = sessionManager

	mailChannel := make(chan models.MailData)
	testApp.MailChannel = mailChannel
	defer close(mailChannel)
	listenForMail()

	testApp.PathToTemplate = "./../../templates"
	testApp.TemplateCache = map[string]*template.Template{}
	testApp.UseCache = true // not need to rebuild template, use template cache for testing

	repo := NewTestRepo(&testApp)
	NewHandlers(repo)

	render.NewRenderer(&testApp)
	helpers.NewHelpers(&testApp)
	/**
	From main.go
	END
	*/

	os.Exit(m.Run())
}

func getRoutes() http.Handler {
	/**
	From routes.go
	START
	*/
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	//mux.Use(NoSurf) // no csrf for testing
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)

	mux.Get("/about", Repo.About)

	mux.Get("/generals-quarters", Repo.Generals)

	mux.Get("/majors-suite", Repo.Majors)

	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)

	mux.Get("/contact", Repo.Contact)

	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostLogin)
	mux.Get("/user/logout", Repo.Logout)

	mux.Get("/admin/dashboard", Repo.AdminDashboard)

	mux.Get("/admin/reservations-new", Repo.AdminNewReservations)
	mux.Get("/admin/reservations-all", Repo.AdminAllReservations)
	mux.Get("/admin/reservations-calendar", Repo.AdminReservationsCalendar)
	mux.Post("/admin/reservations-calendar", Repo.AdminPostReservationsCalendar)
	mux.Get("/admin/process-reservation/{src}/{id}/action", Repo.AdminProcessReservation)
	mux.Get("/admin/delete-reservation/{src}/{id}/action", Repo.AdminDeleteReservation)

	mux.Get("/admin/reservations/{src}/{id}/show", Repo.AdminShowReservation)
	mux.Post("/admin/reservations/{src}/{id}", Repo.AdminPostShowReservation)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	/**
	From routes.go
	END
	*/

	return mux
}

/**
From middleware.go
START
*/
// NoSurf adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",                  // for all paths
		Secure:   testApp.InProduction, // true for https
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return sessionManager.LoadAndSave(next)
}

/**
From middleware.go
END
*/

func listenForMail() {
	go func() {
		for {
			_ = <-testApp.MailChannel
		}
	}()
}
