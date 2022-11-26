package render

import (
	"github.com/loidinhm31/go-bookings-system/internal/models"
	"html/template"
	"net/http"
	"testing"
)

const pathToTemplate = "./../../templates"

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	var valSession = "123"
	sessionManager.Put(r.Context(), "success", valSession)

	result := addDefaultData(&td, r)

	if result.Flash != valSession {
		t.Errorf("success value of %s not found in session", valSession)
	}
}

func TestTemplate(t *testing.T) {
	app.PathToTemplate = "./../../templates"
	app.TemplateCache = map[string]*template.Template{}
	app.UseCache = true

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	var ww testWriter

	err = Template(&ww, r, "home.page.tmpl", &models.TemplateData{})
	if err != nil {
		t.Error("error writing template to browser\n", err)
	}

	err = Template(&ww, r, "non-existent.page.tmpl", &models.TemplateData{})
	if err == nil {
		t.Error("rendered template that does not exist\n", err)
	}
}

func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = sessionManager.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)

	return r, nil
}

func TestNewTemplates(t *testing.T) {
	NewRenderer(app)
}
