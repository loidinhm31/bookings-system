package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/loidinhm31/bookings-system/internal/config"
	"github.com/loidinhm31/bookings-system/internal/constants"
	"github.com/loidinhm31/bookings-system/internal/driver"
	"github.com/loidinhm31/bookings-system/internal/forms"
	"github.com/loidinhm31/bookings-system/internal/models"
	"github.com/loidinhm31/bookings-system/internal/render"
	"github.com/loidinhm31/bookings-system/internal/repository"
	"github.com/loidinhm31/bookings-system/internal/repository/dbrepo"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date""`
	EndDate   string `json:"end_date"`
}

var Repo *Repository

// NewRepo creates a new repository for the handler
func NewRepo(appConfig *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: appConfig,
		DB:  dbrepo.NewPostgresRepo(db.SQL, appConfig),
	}
}

func NewTestRepo(testAppConfig *config.AppConfig) *Repository {
	return &Repository{
		App: testAppConfig,
		DB:  dbrepo.NewTestingRepo(testAppConfig),
	}
}

// NewHandlers sets the repository for the handler
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := m.App.SessionManager.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.SessionManager.Put(r.Context(), "error", "Can't get reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// get room information from db
	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "Can't find room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	res.Room.RoomName = room.RoomName
	m.App.SessionManager.Put(r.Context(), "reservation", res) // update reservation value in the session

	sd := res.StartDate.Format(constants.Layout)
	ed := res.EndDate.Format(constants.Layout)

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "Can't parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation, ok := m.App.SessionManager.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.SessionManager.Put(r.Context(), "error", "Can't get reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")
	reservation.RoomID, err = strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "Can't get room")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// check form valid
	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		http.Error(w, "invalid form", http.StatusSeeOther)
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// insert reservation to db
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "Can't save reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomRestriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	// insert room restriction into db
	err = m.DB.InsertRoomRestriction(roomRestriction)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "Can't save room reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// send mail notifications - guest
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s, <br>
		This is confirm your reservaton from %s to %s.
	`, reservation.FirstName, reservation.StartDate.Format(constants.Layout), reservation.EndDate.Format(constants.Layout))

	msg := models.MailData{
		To:           reservation.Email,
		From:         "me@here.com",
		Subject:      "Reservation Confirmation",
		Content:      htmlMessage,
		TemplateMail: "basic.html",
	}
	m.App.MailChannel <- msg

	// send mail notification top property owner
	htmlMessage = fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		A reservation has bee made for %s from %s to %s.
	`, reservation.Room.RoomName, reservation.StartDate.Format(constants.Layout), reservation.EndDate.Format(constants.Layout))

	msg = models.MailData{
		To:           "me@there.com",
		From:         "me@here.com",
		Subject:      "Reservation Notification",
		Content:      htmlMessage,
		TemplateMail: "basic.html",
	}
	m.App.MailChannel <- msg

	m.App.SessionManager.Put(r.Context(), "reservation", reservation) // store reservation to the session
	m.App.SessionManager.Put(r.Context(), "success", "Submit")        // push success alert

	// redirect to another page, avoid submitting one more time
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "can't parse form!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	startDate, err := time.Parse(constants.Layout, sd)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "can't parse start date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	endDate, err := time.Parse(constants.Layout, ed)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "can't parse end date!")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "can't get availability for rooms")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if len(rooms) == 0 {
		// no availability
		m.App.SessionManager.Put(r.Context(), "error", "No Availability")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	reservation := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.SessionManager.Put(r.Context(), "reservation", reservation)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// can't parse form, so return appropriate json
		resp := jsonResponse{
			OK:      false,
			Message: "Internal Server Error",
		}

		out, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	startDate, _ := time.Parse(constants.Layout, sd)
	endDate, _ := time.Parse(constants.Layout, ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, err := m.DB.SearchAvailabilityByRoomIDAndDates(startDate, endDate, roomID)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to database",
		}

		out, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	resp := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	out, _ := json.MarshalIndent(resp, "", "     ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary displays the reservation summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.SessionManager.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		// only redirect, not need to send error to page
		m.App.ErrorLog.Println("Can't get reservation from session")

		m.App.SessionManager.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.SessionManager.Remove(r.Context(), "reservation") // remove session data for reservation

	data := make(map[string]interface{})
	data["reservation"] = reservation

	// format date to display
	sd := reservation.StartDate.Format(constants.Layout)
	ed := reservation.EndDate.Format(constants.Layout)
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

// ChooseRoom displays list of available rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	// split the URL up by /, and grab the 3rd element
	exploded := strings.Split(r.RequestURI, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation, ok := m.App.SessionManager.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.SessionManager.Put(r.Context(), "error", "Can't get reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation.RoomID = roomID
	m.App.SessionManager.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	startDate, _ := time.Parse(constants.Layout, sd)
	endDate, _ := time.Parse(constants.Layout, ed)

	room, err := m.DB.GetRoomByID(roomID)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "Can't get room from database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	var reservation models.Reservation
	reservation.Room = room
	reservation.RoomID = roomID
	reservation.StartDate = startDate
	reservation.EndDate = endDate

	m.App.SessionManager.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "make-reservation", http.StatusSeeOther)
}
