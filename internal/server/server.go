package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

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
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	inputTokens, err := strconv.Atoi(r.FormValue("input_tokens"))
	if err != nil {
		log.Printf("Invalid input tokens: %v", err)
		http.Error(w, "Invalid input tokens", http.StatusBadRequest)
		return
	}
	outputTokens, err := strconv.Atoi(r.FormValue("output_tokens"))
	if err != nil {
		log.Printf("Invalid output tokens: %v", err)
		http.Error(w, "Invalid output tokens", http.StatusBadRequest)
		return
	}
	e := struct {
		batch        string
		prompt       string
		response     string
		model        string
		inputTokens  int
		outputTokens int
	}{
		batch:        r.FormValue("batch"),
		prompt:       r.FormValue("prompt"),
		response:     r.FormValue("response"),
		model:        r.FormValue("model"),
		inputTokens:  inputTokens,
		outputTokens: outputTokens,
	}

	if err := s.logger.Log(e.batch, e.prompt, e.response, e.model, e.inputTokens, e.outputTokens); err != nil {
		log.Printf("Failed to log entry: %v", err)
		http.Error(w, "Failed to log entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	batches, err := s.logger.GetBatches()
	if err != nil {
		log.Printf("Failed to retrieve batches: %v", err)
		http.Error(w, "Failed to retrieve batches", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' https://cdn.jsdelivr.net; script-src 'self' https://cdn.jsdelivr.net https://cdn.jsdelivr.net/npm/marked/ 'unsafe-eval'")

	tmplPath := filepath.Join(s.templateDir, "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	batchIds := make([]string, len(batches))
	for i, batch := range batches {
		batchIds[i] = batch.ID
	}

	data := struct {
		Batches []string
	}{
		Batches: batchIds,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleBatch(w http.ResponseWriter, r *http.Request) {
	batchId := strings.TrimPrefix(r.URL.Path, "/")
	logs, err := s.logger.GetLogs(batchId)
	if err != nil {
		log.Printf("Failed to retrieve logs: %v", err)
		http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
