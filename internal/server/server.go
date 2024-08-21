package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

func (s *Server) HandleGetBatches(w http.ResponseWriter, r *http.Request) {
	batches, err := s.logger.GetBatches()
	if err != nil {
		log.Printf("Failed to retrieve batches: %v", err)
		http.Error(w, "Failed to retrieve batches", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(batches); err != nil {
		log.Printf("Failed to encode batches: %v", err)
		http.Error(w, "Failed to encode batches", http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	batchId := r.URL.Query().Get("batch")
	logs, err := s.logger.GetLogs(batchId)
	if err != nil {
		log.Printf("Failed to retrieve logs: %v", err)
		http.Error(w, "Failed to retrieve logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (s *Server) HandleGetBatch(w http.ResponseWriter, r *http.Request) {
	batchId := chi.URLParam(r, "batchId")
	batch, err := s.logger.GetBatch(batchId)
	if err != nil {
		log.Printf("Failed to retrieve batch: %v", err)
		http.Error(w, "Failed to retrieve batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batch)
}

func (s *Server) HandleCreateBatch(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	description := r.FormValue("description")

	batchID, err := s.logger.CreateBatch(id, description)
	if err != nil {
		log.Printf("Failed to create batch: %v", err)
		http.Error(w, "Failed to create batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"batch_id": batchID.String()})
}
