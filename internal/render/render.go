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
	"simpleDate": SimpleDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
	"add":        Add,
}

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// Template renders templates using html/template
func Template(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) error {
	var t *template.Template
	var err error
	if app.UseCache && len(app.TemplateCache) > 0 && app.TemplateCache[tmpl] != nil {
		// get requested template cache from the app config
		t = app.TemplateCache[tmpl]
	} else {
		// this is just use for testing, so that we rebuild
		// the cache on every request
		t, err = createTemplateCache("layout/*.layout.tmpl", tmpl)
		if err != nil {
			return err
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

func createTemplateCache(layoutSuffix, pageNameExt string) (*template.Template, error) {
	var t *template.Template

	layouts, err := filepath.Glob(fmt.Sprintf("%s/*%s", app.PathToTemplate, layoutSuffix))
	if err != nil {
		return t, err
	}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*%s", app.PathToTemplate, pageNameExt))
	if err != nil {
		return nil, err
	}

	if len(layouts) > 0 && len(pages) > 0 {
		name := filepath.Base(pages[0])

		filenames := make([]string, 0, len(layouts)+1)
		filenames = append(filenames, pages[0])
		filenames = append(filenames, layouts...)

		t, err = template.New(name).Funcs(functions).ParseFiles(filenames...)
		if err != nil {
			return t, err
		}
		return t, nil
	}
	return nil, errors.New("template file is not available")
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

// SimpleDate returns time in YYYY-MM-DD format
func SimpleDate(t time.Time) string {
	return t.Format(constants.Layout)
}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

// Iterate returns a slice of int, starting at 1, going to count
func Iterate(count int) []int {
	var i int
	var items []int

	for i = 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

func Add(a, b int) int {
	return a + b
}
