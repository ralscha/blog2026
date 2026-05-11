package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Idea struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Pitch         string  `json:"pitch"`
	Stage         string  `json:"stage"`
	HypeScore     int     `json:"hypeScore"`
	CreatedAt     string  `json:"createdAt"`
	LastBoostedAt *string `json:"lastBoostedAt,omitempty"`
}

type CreateIdeaRequest struct {
	Title string `json:"title"`
	Pitch string `json:"pitch"`
	Stage string `json:"stage"`
}

type ListIdeasFilter struct {
	Stages  []string
	Query   string
	MinHype *int
	Limit   int
}

type ListIdeasResponse struct {
	Data struct {
		Ideas []Idea `json:"ideas"`
		Count int    `json:"count"`
	} `json:"data"`
	Metadata struct {
		Filter ListIdeasFilter `json:"filter"`
		Host   string          `json:"host"`
	} `json:"metadata"`
}

type IdeaResponse struct {
	Data struct {
		Idea Idea `json:"idea"`
	} `json:"data"`
}

type MutationResponse struct {
	Data struct {
		Message string `json:"message"`
		Idea    Idea   `json:"idea"`
	} `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Data struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Version string `json:"version"`
	} `json:"data"`
}

type AliveResponse struct {
	Data struct {
		Status string `json:"status"`
	} `json:"data"`
}

type Client struct {
	baseURL    string
	metricsURL string
	http       *http.Client
}

func NewClient(baseURL, metricsPort string, timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		metricsURL: "http://localhost:" + metricsPort,
		http:       &http.Client{Timeout: timeout},
	}
}

func (c *Client) ListIdeas(filter *ListIdeasFilter) (*ListIdeasResponse, error) {
	u, err := url.Parse(c.baseURL + "/ideas")
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	if filter != nil {
		for _, stage := range filter.Stages {
			q.Add("stage", stage)
		}
		if filter.Query != "" {
			q.Set("q", filter.Query)
		}
		if filter.MinHype != nil {
			q.Set("minHype", strconv.Itoa(*filter.MinHype))
		}
		if filter.Limit > 0 {
			q.Set("limit", strconv.Itoa(filter.Limit))
		}
	}
	u.RawQuery = q.Encode()

	var resp ListIdeasResponse
	if err := c.do(http.MethodGet, u.String(), nil, &resp); err != nil {
		return nil, fmt.Errorf("list ideas: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetIdea(id int64) (*IdeaResponse, error) {
	u := fmt.Sprintf("%s/ideas/%d", c.baseURL, id)

	var resp IdeaResponse
	if err := c.do(http.MethodGet, u, nil, &resp); err != nil {
		return nil, fmt.Errorf("get idea %d: %w", id, err)
	}
	return &resp, nil
}

func (c *Client) CreateIdea(req *CreateIdeaRequest) (*MutationResponse, error) {
	u := c.baseURL + "/ideas"

	var resp MutationResponse
	if err := c.do(http.MethodPost, u, req, &resp); err != nil {
		return nil, fmt.Errorf("create idea: %w", err)
	}
	return &resp, nil
}

func (c *Client) BoostIdea(id int64) (*MutationResponse, error) {
	u := fmt.Sprintf("%s/ideas/%d/boost", c.baseURL, id)

	var resp MutationResponse
	if err := c.do(http.MethodPost, u, nil, &resp); err != nil {
		return nil, fmt.Errorf("boost idea %d: %w", id, err)
	}
	return &resp, nil
}

func (c *Client) DeleteIdea(id int64) (bool, error) {
	u := fmt.Sprintf("%s/ideas/%d", c.baseURL, id)

	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return false, fmt.Errorf("delete idea %d: %w", id, err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return false, fmt.Errorf("delete idea %d: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return true, nil
	}

	return false, decodeError(resp)
}

func (c *Client) Health() (*HealthResponse, error) {
	u := c.baseURL + "/.well-known/health"

	var resp HealthResponse
	if err := c.do(http.MethodGet, u, nil, &resp); err != nil {
		return nil, fmt.Errorf("health: %w", err)
	}
	return &resp, nil
}

func (c *Client) Alive() (*AliveResponse, error) {
	u := c.baseURL + "/.well-known/alive"

	var resp AliveResponse
	if err := c.do(http.MethodGet, u, nil, &resp); err != nil {
		return nil, fmt.Errorf("alive: %w", err)
	}
	return &resp, nil
}

func (c *Client) OpenAPI() ([]byte, error) {
	u := c.baseURL + "/.well-known/openapi.json"

	resp, err := c.http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("openapi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) Swagger() (string, error) {
	u := c.baseURL + "/.well-known/swagger"

	resp, err := c.http.Get(u)
	if err != nil {
		return "", fmt.Errorf("swagger: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", decodeError(resp)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("swagger read: %w", err)
	}
	return string(b), nil
}

func (c *Client) Metrics() (string, error) {
	resp, err := c.http.Get(c.metricsURL + "/metrics")
	if err != nil {
		return "", fmt.Errorf("metrics: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("metrics read: %w", err)
	}
	return string(b), nil
}

func (c *Client) do(method, url string, body any, target any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return decodeError(resp)
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

func decodeError(resp *http.Response) error {
	b, _ := io.ReadAll(resp.Body)
	var apiErr ErrorResponse
	if json.Unmarshal(b, &apiErr) == nil && apiErr.Error != "" {
		return fmt.Errorf("%s (HTTP %d): %s", resp.Status, resp.StatusCode, apiErr.Error)
	}
	return fmt.Errorf("%s (HTTP %d)", resp.Status, resp.StatusCode)
}
