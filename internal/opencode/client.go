package opencode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL  string
	password string
	client   *http.Client
}

type MessageRequest struct {
	Parts    []Part `json:"parts"`
	Agent    string `json:"agent,omitempty"`
	Model    string `json:"model,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type Part struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type MessageResponse struct {
	Info  Info   `json:"info"`
	Parts []Part `json:"parts"`
}

type Info struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type Session struct {
	ID string `json:"id"`
}

func NewClient(port, password string) *Client {
	baseURL := fmt.Sprintf("http://127.0.0.1:%s", port)
	return &Client{
		baseURL:  baseURL,
		password: password,
		client:   &http.Client{},
	}
}

func (c *Client) SendMessage(sessionID, text, agent, model, provider string) (*MessageResponse, error) {
	url := fmt.Sprintf("%s/session/%s/message", c.baseURL, sessionID)

	reqBody := MessageRequest{
		Parts: []Part{
			{Type: "text", Text: text},
		},
		Agent:    agent,
		Model:    model,
		Provider: provider,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.password != "" {
		req.SetBasicAuth("opencode", c.password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) CreateSession() (*Session, error) {
	url := fmt.Sprintf("%s/session", c.baseURL)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.password != "" {
		req.SetBasicAuth("opencode", c.password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &session, nil
}

func (c *Client) HealthCheck() error {
	url := fmt.Sprintf("%s/global/health", c.baseURL)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}
