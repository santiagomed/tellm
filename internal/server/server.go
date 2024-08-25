package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

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

func (s *Server) HandleCreateBatch(w http.ResponseWriter, r *http.Request) {
	var b logger.BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		log.Printf("Failed to decode JSON: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	batchID, err := s.logger.CreateBatch(ctx, b)
	if err != nil {
		log.Printf("Failed to create batch: %v", err)
		http.Error(w, "Failed to create batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"batch_id": batchID.String()})
}

func (s *Server) HandleLog(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var e logger.EntryRequest

	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		log.Printf("Failed to decode JSON: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.logger.Log(ctx, e); err != nil {
		log.Printf("Failed to log entry: %v", err)
		http.Error(w, "Failed to log entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) HandleGetBatches(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	batches, err := s.logger.GetBatches(ctx)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logs, err := s.logger.GetLogs(ctx, batchId)
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	batch, err := s.logger.GetBatch(ctx, batchId)
	if err != nil {
		log.Printf("Failed to retrieve batch: %v", err)
		http.Error(w, "Failed to retrieve batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batch)
}
