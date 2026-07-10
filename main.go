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

	// Protected routes — wrapped directly with JWT middleware
	mux.Handle("POST /tickets", middleware.Auth(http.HandlerFunc(handlers.CreateTicket)))
	mux.Handle("GET /tickets", middleware.Auth(http.HandlerFunc(handlers.ListTickets)))
	mux.Handle("GET /tickets/{id}", middleware.Auth(http.HandlerFunc(handlers.GetTicket)))
	mux.Handle("PATCH /tickets/{id}/status", middleware.Auth(http.HandlerFunc(handlers.UpdateTicketStatus)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("server running on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
