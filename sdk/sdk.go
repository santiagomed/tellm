package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Client represents the SDK client for interacting with the llmlog server
type Client struct {
	BaseURL string
}

// NewClient creates a new SDK client with the given base URL
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Prompt    string `json:"prompt"`
	Response  string `json:"response"`
}

// Log sends a log entry to the server
func (c *Client) Log(batch, prompt, response string) error {
	data := url.Values{}
	data.Set("batch", batch)
	data.Set("prompt", prompt)
	data.Set("response", response)

	resp, err := http.PostForm(c.BaseURL+"/log", data)
	if err != nil {
		return fmt.Errorf("failed to send log: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetLogs retrieves all log entries from the server
func (c *Client) GetLogs(batch string) ([]LogEntry, error) {
	resp, err := http.Get(c.BaseURL + "/?batch=" + url.QueryEscape(batch))
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var logs []LogEntry
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	return logs, nil
}
