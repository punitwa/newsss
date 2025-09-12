// Package image provides image processing utilities.
package image

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Processor provides image processing and validation functionality.
type Processor struct {
	client *http.Client
	logger zerolog.Logger
}

// ImageInfo contains metadata about an image.
type ImageInfo struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	IsValid     bool   `json:"is_valid"`
	Error       string `json:"error,omitempty"`
}

// NewProcessor creates a new image processor.
func NewProcessor(timeout time.Duration, logger zerolog.Logger) *Processor {
	return &Processor{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        5,
				MaxIdleConnsPerHost: 2,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		logger: logger.With().Str("component", "image_processor").Logger(),
	}
}

// ValidateImage checks if an image URL is valid and accessible.
func (p *Processor) ValidateImage(ctx context.Context, imageURL string) (*ImageInfo, error) {
	info := &ImageInfo{
		URL: imageURL,
	}
	
	if imageURL == "" {
		info.Error = "empty image URL"
		return info, fmt.Errorf("empty image URL")
	}
	
	// Validate URL format
	if !isValidURL(imageURL) {
		info.Error = "invalid URL format"
		return info, fmt.Errorf("invalid URL format")
	}
	
	// Perform HEAD request to check if image exists and get metadata
	req, err := http.NewRequestWithContext(ctx, "HEAD", imageURL, nil)
	if err != nil {
		info.Error = fmt.Sprintf("failed to create request: %v", err)
		return info, err
	}
	
	// Set headers
	req.Header.Set("User-Agent", "NewsAggregator/1.0 ImageValidator")
	req.Header.Set("Accept", "image/*")
	
	resp, err := p.client.Do(req)
	if err != nil {
		info.Error = fmt.Sprintf("request failed: %v", err)
		return info, err
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		info.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
		return info, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	// Extract metadata from headers
	info.ContentType = resp.Header.Get("Content-Type")
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		// Note: Content-Length parsing would require strconv, keeping it simple for now
		info.Size = 0 // Could parse if needed
	}
	
	// Validate content type
	if !p.isValidImageContentType(info.ContentType) {
		info.Error = fmt.Sprintf("invalid content type: %s", info.ContentType)
		return info, fmt.Errorf("invalid content type: %s", info.ContentType)
	}
	
	info.IsValid = true
	
	p.logger.Debug().
		Str("url", imageURL).
		Str("content_type", info.ContentType).
		Int64("size", info.Size).
		Msg("Image validated successfully")
	
	return info, nil
}

// ValidateImages validates multiple image URLs concurrently.
func (p *Processor) ValidateImages(ctx context.Context, imageURLs []string) []*ImageInfo {
	if len(imageURLs) == 0 {
		return nil
	}
	
	results := make([]*ImageInfo, len(imageURLs))
	
	// Use a semaphore to limit concurrent requests
	semaphore := make(chan struct{}, 5) // Max 5 concurrent validations
	
	// Channel to collect results
	resultChan := make(chan struct {
		index int
		info  *ImageInfo
	}, len(imageURLs))
	
	// Start validation goroutines
	for i, url := range imageURLs {
		go func(index int, imageURL string) {
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore
			
			info, err := p.ValidateImage(ctx, imageURL)
			if err != nil {
				p.logger.Debug().
					Err(err).
					Str("url", imageURL).
					Msg("Image validation failed")
			}
			
			resultChan <- struct {
				index int
				info  *ImageInfo
			}{index: index, info: info}
		}(i, url)
	}
	
	// Collect results
	for i := 0; i < len(imageURLs); i++ {
		result := <-resultChan
		results[result.index] = result.info
	}
	
	return results
}

// FilterValidImages returns only the valid images from a list.
func (p *Processor) FilterValidImages(ctx context.Context, imageURLs []string) []string {
	infos := p.ValidateImages(ctx, imageURLs)
	
	var validImages []string
	for _, info := range infos {
		if info != nil && info.IsValid {
			validImages = append(validImages, info.URL)
		}
	}
	
	return validImages
}

// GetBestImage selects the best image from a list based on various criteria.
func (p *Processor) GetBestImage(ctx context.Context, imageURLs []string) (string, error) {
	if len(imageURLs) == 0 {
		return "", fmt.Errorf("no image URLs provided")
	}
	
	// If only one image, validate and return it
	if len(imageURLs) == 1 {
		info, err := p.ValidateImage(ctx, imageURLs[0])
		if err != nil || !info.IsValid {
			return "", fmt.Errorf("image validation failed: %v", err)
		}
		return imageURLs[0], nil
	}
	
	// Validate all images
	infos := p.ValidateImages(ctx, imageURLs)
	
	// Score images based on various criteria
	bestScore := -1.0
	bestImage := ""
	
	for _, info := range infos {
		if info == nil || !info.IsValid {
			continue
		}
		
		score := p.scoreImage(info)
		if score > bestScore {
			bestScore = score
			bestImage = info.URL
		}
	}
	
	if bestImage == "" {
		return "", fmt.Errorf("no valid images found")
	}
	
	return bestImage, nil
}

// scoreImage assigns a score to an image based on various criteria.
func (p *Processor) scoreImage(info *ImageInfo) float64 {
	score := 0.0
	
	// Base score for valid images
	if info.IsValid {
		score += 1.0
	}
	
	// Score based on content type preference
	switch info.ContentType {
	case "image/jpeg", "image/jpg":
		score += 0.5 // JPEG is widely supported
	case "image/png":
		score += 0.4 // PNG is good for quality
	case "image/webp":
		score += 0.3 // WebP is modern but less supported
	case "image/gif":
		score += 0.2 // GIF is less preferred for articles
	default:
		score += 0.1 // Other formats get minimal score
	}
	
	// Score based on URL characteristics (prefer URLs that suggest main content)
	url := strings.ToLower(info.URL)
	
	// Positive indicators
	contentIndicators := []string{
		"article", "content", "main", "hero", "featured", "banner",
		"large", "big", "full", "original",
	}
	
	for _, indicator := range contentIndicators {
		if strings.Contains(url, indicator) {
			score += 0.2
		}
	}
	
	// Negative indicators (UI elements, small images)
	uiIndicators := []string{
		"icon", "logo", "avatar", "thumb", "small", "mini",
		"button", "nav", "menu", "social", "share", "ad",
	}
	
	for _, indicator := range uiIndicators {
		if strings.Contains(url, indicator) {
			score -= 0.3
		}
	}
	
	// Prefer images with size information (when available)
	if info.Size > 0 {
		// Prefer medium to large images (not too small, not too large)
		if info.Size > 10000 && info.Size < 2000000 { // 10KB to 2MB
			score += 0.1
		}
	}
	
	return score
}

// isValidImageContentType checks if a content type represents a valid image.
func (p *Processor) isValidImageContentType(contentType string) bool {
	if contentType == "" {
		return false
	}
	
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	
	// Remove charset and other parameters
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
		"image/svg+xml",
		"image/tiff",
		"image/x-icon",
	}
	
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	
	return false
}

// GetImageExtensionFromURL extracts the file extension from an image URL.
func GetImageExtensionFromURL(imageURL string) string {
	// Find the last dot in the URL path (before query parameters)
	if idx := strings.Index(imageURL, "?"); idx != -1 {
		imageURL = imageURL[:idx]
	}
	
	if idx := strings.LastIndex(imageURL, "."); idx != -1 {
		ext := strings.ToLower(imageURL[idx+1:])
		
		// Validate it's an image extension
		imageExtensions := []string{
			"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg", "tiff", "ico",
		}
		
		for _, validExt := range imageExtensions {
			if ext == validExt {
				return ext
			}
		}
	}
	
	return ""
}

// IsImageURL checks if a URL likely points to an image based on extension.
func IsImageURL(url string) bool {
	return GetImageExtensionFromURL(url) != ""
}
