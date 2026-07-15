package opengolfapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golf-game-kaffip/internal/domain/course"
)

type Client struct {
	http    *http.Client
	baseURL string
	apiKey  string
	logger  *slog.Logger
}

type ClientInterface interface {
	GetCourse(ctx context.Context, id string) (*CourseResponse, error)
}

func NewClient(apiKey string, logger *slog.Logger) *Client {
	return &Client{
		http:    &http.Client{Timeout: 20 * time.Second},
		baseURL: "https://api.opengolfapi.org/api/v1",
		apiKey:  apiKey,
		logger:  logger,
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string) (*http.Request, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) GetCourse(ctx context.Context, courseID string) (*CourseResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/courses/"+courseID)
	if err != nil {
		return nil, err
	}

	c.logger.Info("external api request", "url", req.URL.String())

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// --- NEW: status code handling ---
	if res.StatusCode == http.StatusNotFound {
		return nil, course.ErrCourseNotFound
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	// ---------------------------------

	var out CourseResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
