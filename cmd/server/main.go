package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/santiagomed/tellm/internal/logger"
	"github.com/santiagomed/tellm/internal/server"
)

func main() {
	l := logger.NewLogger("llm_logs.json")

	// Adjust this path according to your project structure
	templateDir := filepath.Join(".", "internal", "templates")

	s := server.NewServer(l, templateDir)

	http.HandleFunc("/log", s.HandleLog)
	http.HandleFunc("/", s.HandleIndex)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
