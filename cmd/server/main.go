package main

import (
	"log"
	"net/http"

	"tellm/internal/logger"
	"tellm/internal/server"
)

func main() {
	l := logger.NewLogger("llm_logs.json")
	s := server.NewServer(l)

	http.HandleFunc("/log", s.HandleLog)
	http.HandleFunc("/", s.HandleIndex)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
