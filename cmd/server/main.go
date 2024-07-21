package main

import (
	"log"
	"net/http"

	"github.com/santiagomed/tellm/internal/logger"
	"github.com/santiagomed/tellm/internal/server"
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
