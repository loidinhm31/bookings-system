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
)

var app *config.AppConfig
var functions = template.FuncMap{}

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// Template renders templates using html/template
func Template(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) error {
	var templateCache map[string]*template.Template

	if app.UseCache {
		// get the template cache from the app config
		templateCache = app.TemplateCache
	} else {
		// this is just use for testing, so that we rebuild
		// the cache on every request
		templateCache, _ = CreateTemplateCache(constants.PathToTemplate)
	}

	// get requested template from cache
	t, ok := templateCache[tmpl]
	if !ok {
		return errors.New("could not get template from template cache")
	}

	buff := new(bytes.Buffer)

	templateData = addDefaultData(templateData, r)

	_ = t.Execute(buff, templateData)

	// render the template
	_, err := buff.WriteTo(w)
	if err != nil {
		log.Println("Error writing template", err)
		return err
	}
	return nil
}

func CreateTemplateCache(pathToTemplate string) (map[string]*template.Template, error) {
	someCache := map[string]*template.Template{}

	// get all the files with pattern *page.tmpl
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplate))
	if err != nil {
		return someCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return someCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
		if err != nil {
			return someCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
			if err != nil {
				return someCache, err
			}
		}

		someCache[name] = ts
	}
	return someCache, nil
}

func addDefaultData(templateData *models.TemplateData, r *http.Request) *models.TemplateData {
	// Push message in the session for the next time page is displayed
	templateData.Flash = app.SessionManager.PopString(r.Context(), "success")
	templateData.Error = app.SessionManager.PopString(r.Context(), "error")
	templateData.Warning = app.SessionManager.PopString(r.Context(), "warning")

	templateData.CSRFToken = nosurf.Token(r)
	return templateData
}
