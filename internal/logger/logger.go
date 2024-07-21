package logger

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Prompt    string    `json:"prompt"`
	Response  string    `json:"response"`
}

type Logger struct {
	filename string
	mu       sync.Mutex
}

func NewLogger(filename string) *Logger {
	return &Logger{filename: filename}
}

func (l *Logger) Log(prompt, response string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Prompt:    prompt,
		Response:  response,
	}

	file, err := os.OpenFile(l.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(entry)
}

func (l *Logger) GetLogs() ([]LogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	file, err := os.OpenFile(l.filename, os.O_RDONLY, 0644)
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
