package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	Timestamp    string `json:"timestamp"`
	Prompt       string `json:"prompt"`
	Response     string `json:"response"`
	Model        string `json:"model"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// Log sends a log entry to the server
func (c *Client) Log(batch, prompt, response, model string, inputTokens, outputTokens int) error {
	data := url.Values{}
	data.Set("batch", batch)
	data.Set("prompt", prompt)
	data.Set("response", response)
	data.Set("model", model)
	data.Set("input_tokens", strconv.Itoa(inputTokens))
	data.Set("output_tokens", strconv.Itoa(outputTokens))

	resp, err := http.PostForm(c.BaseURL+"/logs", data)
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

	var logs []LogEntry
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
func (c *Client) CreateBatch(id, description string) (string, error) {
	resp, err := http.PostForm(c.BaseURL+"/batches", url.Values{"id": {id}, "description": {description}})
	if err != nil {
		return "", fmt.Errorf("failed to create batch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return id, nil
}
