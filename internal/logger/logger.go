package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Prompt    string    `json:"prompt"`
	Response  string    `json:"response"`
}

type Logger struct {
	outDir  string
	batches []string
	mu      sync.Mutex
}

func NewLogger() *Logger {
	outDir := "logs"
	var batches []string
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		os.Mkdir(outDir, 0755)
		batches = []string{}
	} else {
		files, err := os.ReadDir(outDir)
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			batches = append(batches, file.Name())
		}
	}

	return &Logger{outDir: outDir, batches: batches}
}

func (l *Logger) Log(filename, prompt, response string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	path := filepath.Join(l.outDir, filename)

	entry := LogEntry{
		Timestamp: time.Now(),
		Prompt:    prompt,
		Response:  response,
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(entry)
}

func (l *Logger) GetLogs(filename string) ([]LogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	path := filepath.Join(l.outDir, filename)

	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logs []LogEntry
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var entry LogEntry
		if err := decoder.Decode(&entry); err != nil {
			return nil, err
		}
		logs = append(logs, entry)
	}

	return logs, nil
}

func (l *Logger) GetBatches() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.batches
}
