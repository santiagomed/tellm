package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/santiagomed/tellm/internal/logger"
	"github.com/santiagomed/tellm/internal/server"
)

func main() {
	if os.Getenv("MONGODB_URI") == "" {
		log.Fatal("MONGODB_URI is not set")
	}

	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	// Adjust this path according to your project structure
	templateDir := filepath.Join(".", "internal", "templates")

	s := server.NewServer(l, templateDir)

	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Add CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Routes
	r.Post("/logs", s.HandleLog)
	r.Get("/logs", s.HandleGetLogs)
	r.Get("/batches", s.HandleGetBatches)
	r.Get("/batches/{batchId}", s.HandleGetBatch)
	r.Post("/batches", s.HandleCreateBatch)

	log.Println("Server starting on :8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal(err)
	}
}
