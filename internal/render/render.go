package render

import (
	"bytes"
	"github.com/justinas/nosurf"
	"github.com/loidinhm31/access-system/internal/config"
	"github.com/loidinhm31/access-system/internal/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var app *config.AppConfig

func NewTemplates(a *config.AppConfig) {
	app = a
}

func DrawTemplate(w http.ResponseWriter, r *http.Request, tmpl string, templateData *models.TemplateData) {
	var templateCache map[string]*template.Template

	if app.UseCache {
		// get the t cache from the app config
		templateCache = app.TemplateCache
	} else {
		templateCache, _ = CreateTemplateCache()
	}

	// get requested t from cache
	t, ok := templateCache[tmpl]
	if !ok {
		log.Fatal("Could not get t from t cache")
	}

	buff := new(bytes.Buffer)

	templateData = addDefaultData(templateData, r)

	_ = t.Execute(buff, templateData)

	// render the t
	_, err := buff.WriteTo(w)
	if err != nil {
		log.Println("Error writing t", err)
	}
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	someCache := map[string]*template.Template{}

	// get all the files with pattern *page.tmpl
	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		return someCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return someCache, err
		}

		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			return someCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				return someCache, err
			}
		}

		someCache[name] = ts
	}
	return someCache, nil
}

func addDefaultData(templateData *models.TemplateData, r *http.Request) *models.TemplateData {
	templateData.CSRFToken = nosurf.Token(r)
	return templateData
}
