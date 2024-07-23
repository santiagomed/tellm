package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/santiagomed/tellm/internal/logger"
)

type Server struct {
	logger      *logger.Logger
	templateDir string
}

func NewServer(logger *logger.Logger, templateDir string) *Server {
	return &Server{
		logger:      logger,
		templateDir: templateDir,
	}
}

func (s *Server) HandleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	batch := r.FormValue("batch")
	prompt := r.FormValue("prompt")
	response := r.FormValue("response")

	if err := s.logger.Log(batch, prompt, response); err != nil {
		http.Error(w, "Failed to log entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	batch := r.URL.Query().Get("batch")
	logs, err := s.logger.GetLogs(batch)
	if err != nil {
		http.Error(w, "Failed to retrieve logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Debug: Print the number of logs retrieved
	log.Printf("Retrieved %d logs", len(logs))

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
		return
	}

	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' https://cdn.jsdelivr.net; script-src 'self' https://cdn.jsdelivr.net https://cdn.jsdelivr.net/npm/marked/ 'unsafe-eval'")

	tmplPath := filepath.Join(s.templateDir, "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Failed to parse template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert logs to JSON for Alpine.js
	logsJSON, err := json.Marshal(logs)
	if err != nil {
		http.Error(w, "Failed to marshal logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Logs template.JS
	}{
		Logs: template.JS(logsJSON),
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleBatches(w http.ResponseWriter, r *http.Request) {
	batches := s.logger.GetBatches()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batches)
}
