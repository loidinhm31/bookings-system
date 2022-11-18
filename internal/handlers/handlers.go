package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/loidinhm31/bookings-system/internal/config"
	"github.com/loidinhm31/bookings-system/internal/constants"
	"github.com/loidinhm31/bookings-system/internal/driver"
	"github.com/loidinhm31/bookings-system/internal/forms"
	"github.com/loidinhm31/bookings-system/internal/helpers"
	"github.com/loidinhm31/bookings-system/internal/models"
	"github.com/loidinhm31/bookings-system/internal/render"
	"github.com/loidinhm31/bookings-system/internal/repository"
	"github.com/loidinhm31/bookings-system/internal/repository/dbrepo"
	"log"
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// get room information from db
	room, err := m.DB.GetRoomByID(res.RoomID)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "Can't find room")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	reservation, ok := m.App.SessionManager.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.SessionManager.Put(r.Context(), "error", "Can't get reservation")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")
	reservation.RoomID, err = strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "Can't get room")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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

		stringMap := make(map[string]string)
		stringMap["start_date"] = r.Form.Get("start_date")
		stringMap["end_date"] = r.Form.Get("end_date")

		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form:      form,
			Data:      data,
			StringMap: stringMap,
		})
		return
	}

	// insert reservation to db
	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "Can't save reservation")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	startDate, err := time.Parse(constants.Layout, sd)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "can't parse start date!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	endDate, err := time.Parse(constants.Layout, ed)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "can't parse end date!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		m.App.SessionManager.Put(r.Context(), "error", "can't get availability for rooms")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	reservation, ok := m.App.SessionManager.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.SessionManager.Put(r.Context(), "error", "Can't get reservation")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	var reservation models.Reservation
	reservation.Room = room
	reservation.RoomID = roomID
	reservation.StartDate = startDate
	reservation.EndDate = endDate

	m.App.SessionManager.Put(r.Context(), "reservation", reservation)

	http.Redirect(w, r, "make-reservation", http.StatusSeeOther)
}

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	// ensure renew token
	_ = m.App.SessionManager.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Template(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)
		m.App.SessionManager.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.SessionManager.Put(r.Context(), "user_id", id)
	m.App.SessionManager.Put(r.Context(), "success", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.SessionManager.Destroy(r.Context())
	_ = m.App.SessionManager.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "admin/admin-dashboard.page.tmpl", &models.TemplateData{})
}

func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin/admin-new-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {
	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Template(w, r, "admin/admin-all-reservations.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year

	// get reservation from the db
	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "admin/admin-reservations-show.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      forms.New(nil),
	})
}

func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.SessionManager.Put(r.Context(), "success", "Changes saved")

	month := r.Form.Get("month")
	year := r.Form.Get("year")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	// with no mont/ year specified
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	nextM := now.AddDate(0, 1, 0)
	lastM := now.AddDate(0, -1, 0)

	nextMonth := nextM.Format("01")
	nextMonthYear := nextM.Format("2006")

	lastMonth := lastM.Format("01")
	lastMonthYear := lastM.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _, x := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format(constants.Layout)] = 0
			blockMap[d.Format(constants.Layout)] = 0
		}

		// get all the restrictions for the current room
		roomRestrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range roomRestrictions {
			if y.ReservationID > 0 {
				// it's a reservation
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format(constants.LayoutCalendar)] = y.ReservationID
				}
			} else {
				// it's a block
				blockMap[y.StartDate.Format(constants.LayoutCalendar)] = y.ID
			}
		}
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.SessionManager.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
	}

	render.Template(w, r, "admin/admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

func (m *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	_ = m.DB.UpdateProcessedForReservation(id, 1)

	m.App.SessionManager.Put(r.Context(), "success", "Reservation marked as processed")

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	_ = m.DB.DeleteReservation(id)
	m.App.SessionManager.Put(r.Context(), "success", "Reservation deleted")

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")
	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	// process blocks
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
	}

	form := forms.New(r.PostForm)

	for _, x := range rooms {
		// Get the block map from the session
		// Loop through entire map, if it has an entry in the map that does not exist in the posted data,
		// and if the restriction id > 0, then it is a block need to be removed
		currMap := m.App.SessionManager.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)
		for name, value := range currMap {
			// ok will be false if the value is not in the map
			if val, ok := currMap[name]; ok {
				// only pay attention to values > 0, and that are not in the form post
				// the rest are just placeholders for days without blocks
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
						// delete the restriction by id
						err := m.DB.DeleteBlockRoomRestrictionByID(value)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}

	// handler new blocks
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse(constants.LayoutCalendar, exploded[3])

			// insert a new block
			err := m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				log.Println(err)
			}
		}
	}

	m.App.SessionManager.Put(r.Context(), "success", "Changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}
