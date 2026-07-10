package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"ticket-system/db"
	"ticket-system/handlers"
	"ticket-system/middleware"
)

func main() {
	// Connect to PostgreSQL and run migrations
	db.Connect()

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /auth/register", handlers.Register)
	mux.HandleFunc("POST /auth/login", handlers.Login)

	// Protected routes — wrapped with JWT middleware
	protected := http.NewServeMux()
	protected.HandleFunc("POST /tickets", handlers.CreateTicket)
	protected.HandleFunc("GET /tickets", handlers.ListTickets)
	protected.HandleFunc("GET /tickets/{id}", handlers.GetTicket)
	protected.HandleFunc("PATCH /tickets/{id}/status", handlers.UpdateTicketStatus)

	mux.Handle("/tickets", middleware.Auth(protected))
	mux.Handle("/tickets/", middleware.Auth(protected))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("server running on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
