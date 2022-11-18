package render

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/justinas/nosurf"
	"github.com/loidinhm31/bookings-system/internal/config"
	"github.com/loidinhm31/bookings-system/internal/constants"
	"github.com/loidinhm31/bookings-system/internal/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

var app *config.AppConfig
var functions = template.FuncMap{
	"simpleDate": simpleDate,
	"formatDate": formatDate,
	"iterate":    iterate,
	"add":        add,
}

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// Template renders templates using html/template
func Template(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) error {
	var templateCache map[string]*template.Template

	var t *template.Template
	var err error
	if app.UseCache && app.TemplateCache[tmpl] != nil {
		// get requested template cache from the app config
		t = templateCache[tmpl]
	} else {
		// this is just use for testing, so that we rebuild
		// the cache on every request
		t, err = CreateTemplateCache(constants.PathToTemplate, "layout/*.layout.tmpl", tmpl)
		if err != nil {
			log.Fatal(err)
		}
		app.TemplateCache[tmpl] = t
	}

	buff := new(bytes.Buffer)

	templateData = addDefaultData(templateData, r)

	_ = t.Execute(buff, templateData)

	// render the template
	_, err = buff.WriteTo(w)
	if err != nil {
		log.Println("Error writing template", err)
		return err
	}
	return nil
}

func CreateTemplateCache(pathToTemplate, layoutSuffix, pageNameExt string) (*template.Template, error) {
	var t *template.Template

	layouts, err := filepath.Glob(fmt.Sprintf("%s/*%s", pathToTemplate, layoutSuffix))
	if err != nil {
		return t, err
	}

	page, err := filepath.Glob(fmt.Sprintf("%s/*%s", pathToTemplate, pageNameExt))
	if err != nil {
		if len(page) != 1 {
			return t, errors.New("cannot find template")
		}
		return t, err
	}

	name := filepath.Base(page[0])

	filenames := make([]string, 0, len(layouts)+1)
	filenames = append(filenames, page[0])
	filenames = append(filenames, layouts...)

	t, err = template.New(name).Funcs(functions).ParseFiles(filenames...)
	if err != nil {
		return t, err
	}

	return t, nil
}

func addDefaultData(templateData *models.TemplateData, r *http.Request) *models.TemplateData {
	// Push message in the session for the next time page is displayed
	templateData.Flash = app.SessionManager.PopString(r.Context(), "success")
	templateData.Error = app.SessionManager.PopString(r.Context(), "error")
	templateData.Warning = app.SessionManager.PopString(r.Context(), "warning")

	templateData.CSRFToken = nosurf.Token(r)

	if app.SessionManager.Exists(r.Context(), "user_id") {
		templateData.IsAuthenticated = 1
	}
	return templateData
}

// returns time in YYYY-MM-DD format
func simpleDate(t time.Time) string {
	return t.Format(constants.Layout)
}

func formatDate(t time.Time, f string) string {
	return t.Format(f)
}

// returns a slice of int, starting at 1, going to count
func iterate(count int) []int {
	var i int
	var items []int

	for i = 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

func add(a, b int) int {
	return a + b
}
