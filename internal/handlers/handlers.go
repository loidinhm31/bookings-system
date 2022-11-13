package handlers

import (
	"fmt"
	"github.com/loidinhm31/access-system/internal/config"
	"github.com/loidinhm31/access-system/internal/models"
	"github.com/loidinhm31/access-system/internal/render"
	"net/http"
)

type Repository struct {
	App *config.AppConfig
}

var Repo *Repository

// CreateNewRepo creates a new repository for the handler
func CreateNewRepo(appConfig *config.AppConfig) *Repository {
	return &Repository{App: appConfig}
}

// SetRepository sets the repository for the handler
func SetRepository(r *Repository) {
	Repo = r
}

func (m *Repository) Home(writer http.ResponseWriter, request *http.Request) {
	remoteIp := request.RemoteAddr

	m.App.SessionManager.Put(request.Context(), "remote_ip", remoteIp)

	render.DrawTemplate(writer, request, "home.page.tmpl", &models.TemplateData{})
}

func (m *Repository) About(writer http.ResponseWriter, request *http.Request) {
	// perform some logic
	stringMap := make(map[string]string)

	remoteIp := m.App.SessionManager.GetString(request.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIp

	render.DrawTemplate(writer, request, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) Reservation(writer http.ResponseWriter, request *http.Request) {

	render.DrawTemplate(writer, request, "make-reservation.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Generals(writer http.ResponseWriter, request *http.Request) {

	render.DrawTemplate(writer, request, "generals.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Majors(writer http.ResponseWriter, request *http.Request) {

	render.DrawTemplate(writer, request, "majors.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Availability(writer http.ResponseWriter, request *http.Request) {

	render.DrawTemplate(writer, request, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) PostAvailability(writer http.ResponseWriter, request *http.Request) {
	start := request.Form.Get("start")
	end := request.Form.Get("end")

	writer.Write([]byte(fmt.Sprintf("Start date is %s and and date is %s", start, end)))
}

func (m *Repository) Contact(writer http.ResponseWriter, request *http.Request) {

	render.DrawTemplate(writer, request, "contact.page.tmpl", &models.TemplateData{})
}
