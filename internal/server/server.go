package server

import (
	"encoding/json"
	"html/template"
	"net/http"

	"tellm/internal/logger"
)

type Server struct {
	logger *logger.Logger
}

func NewServer(logger *logger.Logger) *Server {
	return &Server{logger: logger}
}

func (s *Server) HandleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prompt := r.FormValue("prompt")
	response := r.FormValue("response")

	if err := s.logger.Log(prompt, response); err != nil {
		http.Error(w, "Failed to log entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	logs, err := s.logger.GetLogs()
	if err != nil {
		http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logs)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, logs); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
