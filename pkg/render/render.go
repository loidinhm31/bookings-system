package render

import (
	"bytes"
	"github.com/loidinhm31/access-system/pkg/config"
	"github.com/loidinhm31/access-system/pkg/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var app *config.AppConfig

func NewTemplates(a *config.AppConfig) {
	app = a
}

func DrawTemplate(w http.ResponseWriter, tmpl string, templateData *models.TemplateData) {
	var templateCache map[string]*template.Template

	if app.UseCache {
		// get the template cache from the app config
		templateCache = app.TemplateCache
	} else {
		templateCache, _ = CreateTemplateCache()
	}

	// get requested template from cache
	template, ok := templateCache[tmpl]
	if !ok {
		log.Fatal("Could not get template from template cache")
	}

	buff := new(bytes.Buffer)

	templateData = addDefaultData(templateData)

	_ = template.Execute(buff, templateData)

	// render the template
	_, err := buff.WriteTo(w)
	if err != nil {
		log.Println("Error writing template", err)
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

func addDefaultData(templateData *models.TemplateData) *models.TemplateData {

	return templateData
}
