package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/santiagomed/tellm/internal/logger"
	"github.com/santiagomed/tellm/internal/server"
)

func main() {
	l, err := logger.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	// Adjust this path according to your project structure
	templateDir := filepath.Join(".", "internal", "templates")

	s := server.NewServer(l, templateDir)

	http.HandleFunc("/log", s.HandleLog)
	http.HandleFunc("/", s.HandleIndex)
	http.HandleFunc("/{batchId}", s.HandleBatch) // Add this line

	log.Println("Server starting on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
