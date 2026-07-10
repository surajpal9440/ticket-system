package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"ticket-system/db"
	"ticket-system/middleware"
	"ticket-system/models"

	"github.com/google/uuid"
)

type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

// validTransitions defines the only allowed forward moves
var validTransitions = map[string]string{
	"open":        "in_progress",
	"in_progress": "closed",
}

func CreateTicket(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	_, err := db.DB.Exec(
		`INSERT INTO tickets (id, user_id, title, description, status, created_at)
		 VALUES (?, ?, ?, ?, 'open', ?)`,
		id, userID, req.Title, req.Description, now.Format(time.RFC3339),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create ticket")
		return
	}

	ticket := models.Ticket{
		ID:          id,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "open",
		CreatedAt:   now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

func ListTickets(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	rows, err := db.DB.Query(
		`SELECT id, user_id, title, description, status, created_at
		 FROM tickets WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch tickets")
		return
	}
	defer rows.Close()

	tickets := []models.Ticket{}
	for rows.Next() {
		var t models.Ticket
		var createdAt string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &createdAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan ticket")
			return
		}
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		tickets = append(tickets, t)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickets)
}

func GetTicket(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	ticketID := r.PathValue("id")

	var ticket models.Ticket
	var createdAt string
	err := db.DB.QueryRow(
		`SELECT id, user_id, title, description, status, created_at
		 FROM tickets WHERE id = ? AND user_id = ?`,
		ticketID, userID,
	).Scan(&ticket.ID, &ticket.UserID, &ticket.Title, &ticket.Description, &ticket.Status, &createdAt)

	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch ticket")
		return
	}
	ticket.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ticket)
}

func UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	ticketID := r.PathValue("id")

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Fetch current ticket (ownership enforced via user_id)
	var ticket models.Ticket
	var createdAt string
	err := db.DB.QueryRow(
		`SELECT id, user_id, title, description, status, created_at
		 FROM tickets WHERE id = ? AND user_id = ?`,
		ticketID, userID,
	).Scan(&ticket.ID, &ticket.UserID, &ticket.Title, &ticket.Description, &ticket.Status, &createdAt)

	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch ticket")
		return
	}

	// closed is terminal — cannot be moved back
	if ticket.Status == "closed" {
		writeError(w, http.StatusBadRequest, "closed ticket cannot be reopened")
		return
	}

	// Validate the transition
	allowedNext, ok := validTransitions[ticket.Status]
	if !ok || allowedNext != req.Status {
		writeError(w, http.StatusBadRequest, "invalid status transition: "+ticket.Status+" -> "+req.Status)
		return
	}

	// Apply the update
	_, err = db.DB.Exec(
		`UPDATE tickets SET status = ? WHERE id = ? AND user_id = ?`,
		req.Status, ticketID, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update ticket")
		return
	}

	ticket.Status = req.Status
	ticket.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ticket)
}
