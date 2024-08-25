package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/santiagomed/tellm/internal/logger"
)

type LogEntry = logger.LogEntry
type EntryRequest = logger.EntryRequest
type BatchRequest = logger.BatchRequest

// Client represents the SDK client for interacting with the llmlog server
type Client struct {
	BaseURL string
}

// NewClient creates a new SDK client with the given base URL
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

// Log sends a log entry to the server
func (c *Client) Log(req EntryRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.BaseURL+"/logs", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send log: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetLogs retrieves all log entries for a specific batch from the server
func (c *Client) GetLogs(batch string) ([]LogEntry, error) {
	resp, err := http.Get(c.BaseURL + "/" + url.PathEscape(batch))
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var logs []logger.LogEntry
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&logs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %w", err)
	}

	return logs, nil
}

// GetBatches retrieves all available batch IDs from the server
func (c *Client) GetBatches() ([]string, error) {
	resp, err := http.Get(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get batches: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data struct {
		Batches []string `json:"batches"`
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode batches: %w", err)
	}

	return data.Batches, nil
}

// CreateBatch creates a new batch on the server
func (c *Client) CreateBatch(req BatchRequest) (string, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(c.BaseURL+"/batches", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create batch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return req.ID, nil
}
