// Package utils provides common HTTP utilities for data sources.
package utils

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"news-aggregator/internal/datasources/core"

	"github.com/rs/zerolog"
)

// HTTPClient provides HTTP functionality for data sources.
type HTTPClient struct {
	client    *http.Client
	userAgent string
	logger    zerolog.Logger
}

// NewHTTPClient creates a new HTTP client with the specified configuration.
func NewHTTPClient(timeout time.Duration, userAgent string, logger zerolog.Logger) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		userAgent: userAgent,
		logger:    logger.With().Str("component", "http_client").Logger(),
	}
}

// Get performs a GET request with the specified headers.
func (hc *HTTPClient) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set default headers
	req.Header.Set("User-Agent", hc.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	
	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	hc.logger.Debug().
		Str("url", url).
		Str("method", "GET").
		Msg("Making HTTP request")
	
	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, core.NewSourceError("http_client", core.SourceTypeAPI, "request", err)
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, core.NewSourceErrorWithCode("http_client", core.SourceTypeAPI, "request", 
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status), resp.StatusCode)
	}
	
	// Handle response body
	var reader io.Reader = resp.Body
	
	// Handle gzip encoding
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}
	
	// Read response body
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	hc.logger.Debug().
		Str("url", url).
		Int("status_code", resp.StatusCode).
		Int("content_length", len(body)).
		Msg("HTTP request completed")
	
	return body, nil
}

// Post performs a POST request with the specified body and headers.
func (hc *HTTPClient) Post(ctx context.Context, url string, body []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set default headers
	req.Header.Set("User-Agent", hc.userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	hc.logger.Debug().
		Str("url", url).
		Str("method", "POST").
		Int("body_length", len(body)).
		Msg("Making HTTP request")
	
	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, core.NewSourceError("http_client", core.SourceTypeAPI, "request", err)
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, core.NewSourceErrorWithCode("http_client", core.SourceTypeAPI, "request",
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status), resp.StatusCode)
	}
	
	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	hc.logger.Debug().
		Str("url", url).
		Int("status_code", resp.StatusCode).
		Int("response_length", len(responseBody)).
		Msg("HTTP request completed")
	
	return responseBody, nil
}

// SetTimeout updates the HTTP client timeout.
func (hc *HTTPClient) SetTimeout(timeout time.Duration) {
	hc.client.Timeout = timeout
}

// SetUserAgent updates the User-Agent header.
func (hc *HTTPClient) SetUserAgent(userAgent string) {
	hc.userAgent = userAgent
}

// Head performs a HEAD request to check if a resource exists.
func (hc *HTTPClient) Head(ctx context.Context, url string, headers map[string]string) error {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create HEAD request: %w", err)
	}
	
	// Set headers
	req.Header.Set("User-Agent", hc.userAgent)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	resp, err := hc.client.Do(req)
	if err != nil {
		return core.NewSourceError("http_client", core.SourceTypeAPI, "head_request", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return core.NewSourceErrorWithCode("http_client", core.SourceTypeAPI, "head_request",
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status), resp.StatusCode)
	}
	
	return nil
}

// GetContentType extracts the content type from the response headers.
func GetContentType(resp *http.Response) core.ContentType {
	contentType := resp.Header.Get("Content-Type")
	
	switch {
	case strings.Contains(contentType, "application/json"):
		return core.ContentTypeJSON
	case strings.Contains(contentType, "application/xml"), strings.Contains(contentType, "text/xml"):
		return core.ContentTypeXML
	case strings.Contains(contentType, "text/html"):
		return core.ContentTypeHTML
	case strings.Contains(contentType, "text/plain"):
		return core.ContentTypeText
	default:
		return core.ContentTypeText
	}
}

// IsValidURL performs basic URL validation.
func IsValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// SanitizeURL cleans and validates a URL.
func SanitizeURL(url string) string {
	url = strings.TrimSpace(url)
	if !IsValidURL(url) {
		return ""
	}
	return url
}
