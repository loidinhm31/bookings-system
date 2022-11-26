package render

import (
	"encoding/gob"
	"github.com/alexedwards/scs/v2"
	"github.com/loidinhm31/go-bookings-system/internal/config"
	"github.com/loidinhm31/go-bookings-system/internal/models"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var sessionManager *scs.SessionManager
var testApp config.AppConfig

type testWriter struct {
}

func TestMain(m *testing.M) {
	/**
	From main.go
	START
	*/
	// Values using in the session
	gob.Register(models.Reservation{})

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
	sessionManager.Cookie.Secure = false

	testApp.SessionManager = sessionManager
	/**
	From main.go
	END
	*/

	app = &testApp
	os.Exit(m.Run())
}

func (tw *testWriter) Header() http.Header {
	var h http.Header
	return h
}

func (tw *testWriter) WriteHeader(i int) {

}

func (tw *testWriter) Write(b []byte) (int, error) {
	length := len(b)
	return length, nil
}
