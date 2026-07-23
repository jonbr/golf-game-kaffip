package opengolfapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"golf-game-kaffip/internal/domain/course"
)

type Client struct {
	http        *http.Client
	baseURL     string // full course detail: /api/v1
	liteBaseURL string // search + lite lookups: /v1
	apiKey      string
	logger      *slog.Logger
}

type ClientInterface interface {
	GetCourse(ctx context.Context, id string) (*CourseResponse, error)
	SearchCourses(ctx context.Context, query string) ([]CourseSearchResult, error)
}

func NewClient(apiKey string, logger *slog.Logger) *Client {
	return &Client{
		http:        &http.Client{Timeout: 20 * time.Second},
		baseURL:     "https://api.opengolfapi.org/api/v1",
		liteBaseURL: "https://api.opengolfapi.org/v1",
		apiKey:      apiKey,
		logger:      logger,
	}
}

func (c *Client) newRequest(ctx context.Context, method, apiUrl, path string) (*http.Request, error) {
	url := apiUrl + path
	fmt.Println("url:", url)
	c.logger.Info("search auth check", "api_key_length", len(c.apiKey))
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	//spew.Dump(req)

	return req, nil
}

func (c *Client) GetCourse(ctx context.Context, courseID string) (*CourseResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, c.baseURL, "/courses/"+courseID)
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

func (c *Client) SearchCourses(ctx context.Context, query string) ([]CourseSearchResult, error) {
	req, err := c.newRequest(ctx, http.MethodGet, c.liteBaseURL, "/courses/search?q="+url.QueryEscape(query))
	if err != nil {
		return nil, err
	}

	c.logger.Info("external api request", "url", req.URL.String())

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	c.logger.Info("search raw body", "len", len(bodyBytes), "body", string(bodyBytes))

	var out CourseSearchResponse
	if err := json.Unmarshal(bodyBytes, &out); err != nil {
		return nil, err
	}

	return out.Courses, nil
}
