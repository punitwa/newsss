// Package image provides image extraction and processing utilities.
package image

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"news-aggregator/internal/datasources/core"

	"github.com/rs/zerolog"
)

// Scraper provides functionality to extract images from web content.
type Scraper struct {
	client    *http.Client
	userAgent string
	logger    zerolog.Logger
}

// NewScraper creates a new image scraper.
func NewScraper(timeout time.Duration, userAgent string, logger zerolog.Logger) *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        5,
				MaxIdleConnsPerHost: 2,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		userAgent: userAgent,
		logger:    logger.With().Str("component", "image_scraper").Logger(),
	}
}

// ExtractFromURL fetches a webpage and extracts the first valid image.
func (s *Scraper) ExtractFromURL(ctx context.Context, pageURL string) (string, error) {
	if pageURL == "" {
		return "", core.NewValidationError("url", pageURL, "empty article URL")
	}

	// Validate URL
	if !isValidURL(pageURL) {
		return "", core.NewValidationError("url", pageURL, "invalid URL format")
	}

	s.logger.Debug().Str("url", pageURL).Msg("Starting image extraction")

	// Fetch webpage content
	content, err := s.fetchContent(ctx, pageURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch content: %w", err)
	}

	// Extract images from content
	images := s.ExtractFromHTML(content, pageURL)

	// Return the first valid image
	for _, imgURL := range images {
		if s.isValidImageURL(imgURL) {
			s.logger.Debug().
				Str("page_url", pageURL).
				Str("image_url", imgURL).
				Msg("Image extracted successfully")
			return imgURL, nil
		}
	}

	s.logger.Debug().Str("url", pageURL).Msg("No valid images found")
	return "", core.ErrNoContent
}

// ExtractFromHTML extracts image URLs from HTML content.
func (s *Scraper) ExtractFromHTML(htmlContent, baseURL string) []string {
	var images []string

	// Parse base URL for resolving relative URLs
	base, err := url.Parse(baseURL)
	if err != nil {
		s.logger.Error().Err(err).Str("base_url", baseURL).Msg("Failed to parse base URL")
		return images
	}

	// Try different extraction strategies in order of preference
	extractors := []func(string, *url.URL) []string{
		s.extractOpenGraphImages,
		s.extractTwitterCardImages,
		s.extractMetaImages,
		s.extractImgTags,
		s.extractBackgroundImages,
	}

	for i, extractor := range extractors {
		extracted := extractor(htmlContent, base)
		images = append(images, extracted...)

		// If we found images with high-priority extractors (OpenGraph or Twitter), prefer those
		if len(extracted) > 0 && i < 2 {
			break
		}
	}

	// Remove duplicates and invalid URLs
	return s.deduplicateAndValidate(images)
}

// fetchContent retrieves the HTML content from a URL.
func (s *Scraper) fetchContent(ctx context.Context, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set browser-like headers
	s.setBrowserHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", core.NewSourceError("image_scraper", core.SourceTypeScraper, "fetch", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", core.NewSourceErrorWithCode("image_scraper", core.SourceTypeScraper, "fetch",
			fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status), resp.StatusCode)
	}

	// Handle response body
	var reader io.Reader = resp.Body

	// Handle gzip encoding
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Read content
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read content: %w", err)
	}

	return string(content), nil
}

// setBrowserHeaders sets headers to mimic a real browser.
func (s *Scraper) setBrowserHeaders(req *http.Request) {
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "max-age=0")
}

// extractOpenGraphImages extracts Open Graph image URLs.
func (s *Scraper) extractOpenGraphImages(html string, base *url.URL) []string {
	var images []string

	// Look for og:image meta tags
	ogImageRegex := regexp.MustCompile(`<meta\s+(?:property|name)=["']og:image["']\s+content=["']([^"']+)["'][^>]*>`)
	matches := ogImageRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			imgURL := s.resolveURL(match[1], base)
			if imgURL != "" {
				images = append(images, imgURL)
			}
		}
	}

	return images
}

// extractTwitterCardImages extracts Twitter Card image URLs.
func (s *Scraper) extractTwitterCardImages(html string, base *url.URL) []string {
	var images []string

	// Look for twitter:image meta tags
	twitterImageRegex := regexp.MustCompile(`<meta\s+(?:property|name)=["']twitter:image(?::src)?["']\s+content=["']([^"']+)["'][^>]*>`)
	matches := twitterImageRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			imgURL := s.resolveURL(match[1], base)
			if imgURL != "" {
				images = append(images, imgURL)
			}
		}
	}

	return images
}

// extractMetaImages extracts images from other meta tags.
func (s *Scraper) extractMetaImages(html string, base *url.URL) []string {
	var images []string

	// Look for other image meta tags
	patterns := []string{
		`<meta\s+(?:property|name)=["'](?:image|thumbnail)["']\s+content=["']([^"']+)["'][^>]*>`,
		`<link\s+rel=["'](?:image_src|thumbnail)["']\s+href=["']([^"']+)["'][^>]*>`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(html, -1)

		for _, match := range matches {
			if len(match) > 1 {
				imgURL := s.resolveURL(match[1], base)
				if imgURL != "" {
					images = append(images, imgURL)
				}
			}
		}
	}

	return images
}

// extractImgTags extracts images from img tags.
func (s *Scraper) extractImgTags(html string, base *url.URL) []string {
	var images []string

	// Look for img tags with src attribute
	imgTagRegex := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["'][^>]*>`)
	matches := imgTagRegex.FindAllStringSubmatch(html, -1)

	// Limit to first few images to avoid too many results
	maxImages := 10
	count := 0

	for _, match := range matches {
		if count >= maxImages {
			break
		}

		if len(match) > 1 {
			imgURL := s.resolveURL(match[1], base)
			if imgURL != "" && s.isContentImage(match[0]) {
				images = append(images, imgURL)
				count++
			}
		}
	}

	return images
}

// extractBackgroundImages extracts images from CSS background-image properties.
func (s *Scraper) extractBackgroundImages(html string, base *url.URL) []string {
	var images []string

	// Look for background-image in style attributes
	bgImageRegex := regexp.MustCompile(`background-image:\s*url\(["']?([^"')]+)["']?\)`)
	matches := bgImageRegex.FindAllStringSubmatch(html, -1)

	// Limit results
	maxImages := 5
	count := 0

	for _, match := range matches {
		if count >= maxImages {
			break
		}

		if len(match) > 1 {
			imgURL := s.resolveURL(match[1], base)
			if imgURL != "" {
				images = append(images, imgURL)
				count++
			}
		}
	}

	return images
}

// isContentImage checks if an img tag likely contains content (not UI elements).
func (s *Scraper) isContentImage(imgTag string) bool {
	// Skip small images (likely icons or UI elements)
	if strings.Contains(imgTag, `width="`) || strings.Contains(imgTag, `height="`) {
		widthRegex := regexp.MustCompile(`width=["'](\d+)["']`)
		heightRegex := regexp.MustCompile(`height=["'](\d+)["']`)

		if matches := widthRegex.FindStringSubmatch(imgTag); len(matches) > 1 {
			if len(matches[1]) <= 2 { // Width <= 99 pixels
				return false
			}
		}

		if matches := heightRegex.FindStringSubmatch(imgTag); len(matches) > 1 {
			if len(matches[1]) <= 2 { // Height <= 99 pixels
				return false
			}
		}
	}

	// Skip images with certain class names or IDs that suggest UI elements
	uiPatterns := []string{
		"icon", "logo", "avatar", "thumbnail", "button", "nav", "menu", "header", "footer",
		"sidebar", "widget", "ad", "banner", "social", "share",
	}

	imgTagLower := strings.ToLower(imgTag)
	for _, pattern := range uiPatterns {
		if strings.Contains(imgTagLower, pattern) {
			return false
		}
	}

	return true
}

// resolveURL resolves a potentially relative URL against a base URL.
func (s *Scraper) resolveURL(imgURL string, base *url.URL) string {
	if imgURL == "" {
		return ""
	}

	// Parse the image URL
	parsed, err := url.Parse(imgURL)
	if err != nil {
		return ""
	}

	// Resolve against base URL
	resolved := base.ResolveReference(parsed)
	return resolved.String()
}

// isValidImageURL checks if a URL points to a valid image.
func (s *Scraper) isValidImageURL(imgURL string) bool {
	if imgURL == "" {
		return false
	}

	// Check if URL is valid
	if !isValidURL(imgURL) {
		return false
	}

	// Check file extension
	lowerURL := strings.ToLower(imgURL)
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"}

	for _, ext := range imageExtensions {
		if strings.Contains(lowerURL, ext) {
			return true
		}
	}

	// If no extension, it might still be an image (some URLs don't have extensions)
	return true
}

// deduplicateAndValidate removes duplicate and invalid URLs.
func (s *Scraper) deduplicateAndValidate(images []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, img := range images {
		if img != "" && !seen[img] && s.isValidImageURL(img) {
			seen[img] = true
			result = append(result, img)
		}
	}

	return result
}

// isValidURL performs basic URL validation.
func isValidURL(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
