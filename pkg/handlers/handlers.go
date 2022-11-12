package handlers

import (
	"fmt"
	"github.com/loidinhm31/access-system/pkg/config"
	"github.com/loidinhm31/access-system/pkg/models"
	"github.com/loidinhm31/access-system/pkg/render"
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
	fmt.Println(remoteIp)
	m.App.SessionManager.Put(request.Context(), "remote_ip", remoteIp)

	render.DrawTemplate(writer, "home.page.tmpl", &models.TemplateData{})
}

func (m *Repository) About(writer http.ResponseWriter, request *http.Request) {
	// perform some logic
	stringMap := make(map[string]string)

	remoteIp := m.App.SessionManager.GetString(request.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIp

	render.DrawTemplate(writer, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}
