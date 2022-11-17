package dbrepo

import (
	"context"
	"errors"
	"github.com/loidinhm31/bookings-system/internal/models"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (m *postgresDbRepo) AllUsers() bool {
	return true
}

func (m *postgresDbRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int

	stmt := `INSERT INTO reservations (first_name, last_name, email, phone, 
            start_date, end_date, room_id, created_at, updated_at) 
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

func (m *postgresDbRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO room_restrictions(start_date, end_date, room_id, reservation_id,
            created_at, updated_at, restriction_id)
            VALUES ($1, $2, $3, $4, $5, $6, $7);`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)

	if err != nil {
		return err
	}
	return nil
}

// SearchAvailabilityByRoomIDAndDates returns true if availability exists for roomID, and false if no availability
func (m *postgresDbRepo) SearchAvailabilityByRoomIDAndDates(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query := `SELECT count(id) 
			FROM room_restrictions 
			WHERE room_id = $1 
			  AND end_date > $2 
			  AND start_date < $3;`

	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	// check override date range
	if numRows == 0 {
		return true, nil
	}
	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms, if any, for given date range
func (m *postgresDbRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `SELECT r.id, r.room_name 
			FROM rooms r 
			WHERE r.id NOT IN (
			    SELECT room_id FROM room_restrictions rr 
			    WHERE rr.end_date > $1 
					AND rr.start_date < $2
			)`

	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}
	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

func (m *postgresDbRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `SELECT * FROM rooms where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err != nil {
		return room, err
	}
	return room, nil
}

func (m *postgresDbRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT u.* FROM users u WHERE id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (m *postgresDbRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `UPDATE users 
			SET first_name = $1, 
			    last_name = $2,
			    email = $3,
			    access_level = $4,
			    updated_at = $5`

	_, err := m.DB.ExecContext(ctx, stmt,
		u.FirstName,
		u.LastName,
		u.AccessLevel,
		time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDbRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	query := `SELECT u.id, u.password FROM users u WHERE u.email = $1`

	row := m.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}
	return id, hashedPassword, nil
}
