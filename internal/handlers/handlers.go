package handlers

import (
	"fmt"
	"github.com/loidinhm31/access-system/internal/config"
	"github.com/loidinhm31/access-system/internal/forms"
	"github.com/loidinhm31/access-system/internal/models"
	"github.com/loidinhm31/access-system/internal/render"
	"log"
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
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	render.DrawTemplate(writer, request, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostReservation(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	reservation := models.Reservation{
		FirstName: request.Form.Get("first_name"),
		LastName:  request.Form.Get("last_name"),
		Phone:     request.Form.Get("phone"),
		Email:     request.Form.Get("email"),
	}

	form := forms.New(request.PostForm)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.DrawTemplate(writer, request, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
	}

	m.App.SessionManager.Put(request.Context(), "reservation", reservation) // store reservation to the session
	m.App.SessionManager.Put(request.Context(), "success", "Submit")        // push success alert

	// redirect to another page, avoid submitting one more time
	http.Redirect(writer, request, "/reservation-summary", http.StatusSeeOther)
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

func (m *Repository) ReservationSummary(writer http.ResponseWriter, request *http.Request) {
	reservation, ok := m.App.SessionManager.Get(request.Context(), "reservation").(models.Reservation)
	if !ok {
		log.Println("Cannot get item from session")
		m.App.SessionManager.Put(request.Context(), "error", "Can't get reservation from session")
		http.Redirect(writer, request, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.SessionManager.Remove(request.Context(), "reservation") // remove session data for reservation
	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.DrawTemplate(writer, request, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
