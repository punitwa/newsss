// Package rss provides RSS feed parsing and processing functionality.
package rss

import (
	"encoding/xml"
	"time"
)

// Feed represents an RSS feed structure.
type Feed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

// Channel represents an RSS channel.
type Channel struct {
	Title          string `xml:"title"`
	Description    string `xml:"description"`
	Link           string `xml:"link"`
	Language       string `xml:"language,omitempty"`
	Copyright      string `xml:"copyright,omitempty"`
	ManagingEditor string `xml:"managingEditor,omitempty"`
	WebMaster      string `xml:"webMaster,omitempty"`
	PubDate        string `xml:"pubDate,omitempty"`
	LastBuildDate  string `xml:"lastBuildDate,omitempty"`
	Category       string `xml:"category,omitempty"`
	Generator      string `xml:"generator,omitempty"`
	TTL            int    `xml:"ttl,omitempty"`
	Image          *Image `xml:"image,omitempty"`
	Items          []Item `xml:"item"`
}

// Item represents an RSS item.
type Item struct {
	Title       string         `xml:"title"`
	Description string         `xml:"description"`
	Content     string         `xml:"content:encoded"`
	Link        string         `xml:"link"`
	GUID        *GUID          `xml:"guid"`
	PubDate     string         `xml:"pubDate"`
	Author      string         `xml:"author"`
	Category    []Category     `xml:"category"`
	Comments    string         `xml:"comments,omitempty"`
	Enclosure   *Enclosure     `xml:"enclosure,omitempty"`
	Source      *ChannelSource `xml:"source,omitempty"`

	// Dublin Core extensions
	DCCreator string `xml:"http://purl.org/dc/elements/1.1/ creator,omitempty"`
	DCDate    string `xml:"http://purl.org/dc/elements/1.1/ date,omitempty"`
	DCSubject string `xml:"http://purl.org/dc/elements/1.1/ subject,omitempty"`

	// Media RSS extensions
	MediaContent     []MediaContent   `xml:"http://search.yahoo.com/mrss/ content,omitempty"`
	MediaThumbnail   []MediaThumbnail `xml:"http://search.yahoo.com/mrss/ thumbnail,omitempty"`
	MediaDescription string           `xml:"http://search.yahoo.com/mrss/ description,omitempty"`
}

// GUID represents an RSS item GUID.
type GUID struct {
	Value       string `xml:",chardata"`
	IsPermaLink bool   `xml:"isPermaLink,attr,omitempty"`
}

// Category represents an RSS category.
type Category struct {
	Value  string `xml:",chardata"`
	Domain string `xml:"domain,attr,omitempty"`
}

// Enclosure represents an RSS enclosure (typically for media files).
type Enclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length int64  `xml:"length,attr,omitempty"`
}

// ChannelSource represents an RSS channel source element.
type ChannelSource struct {
	Value string `xml:",chardata"`
	URL   string `xml:"url,attr"`
}

// Image represents an RSS channel image.
type Image struct {
	URL         string `xml:"url"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Width       int    `xml:"width,omitempty"`
	Height      int    `xml:"height,omitempty"`
	Description string `xml:"description,omitempty"`
}

// MediaContent represents Media RSS content.
type MediaContent struct {
	URL      string `xml:"url,attr"`
	Type     string `xml:"type,attr,omitempty"`
	Medium   string `xml:"medium,attr,omitempty"`
	Width    int    `xml:"width,attr,omitempty"`
	Height   int    `xml:"height,attr,omitempty"`
	Duration int    `xml:"duration,attr,omitempty"`
	FileSize int64  `xml:"fileSize,attr,omitempty"`
}

// MediaThumbnail represents Media RSS thumbnail.
type MediaThumbnail struct {
	URL    string `xml:"url,attr"`
	Width  int    `xml:"width,attr,omitempty"`
	Height int    `xml:"height,attr,omitempty"`
}

// ParsedItem represents a processed RSS item with normalized fields.
type ParsedItem struct {
	ID          string
	Title       string
	Description string
	Content     string
	Link        string
	Author      string
	PublishedAt time.Time
	Categories  []string
	ImageURL    string
	SourceName  string
	SourceURL   string

	// Additional metadata
	GUID         string
	Comments     string
	EnclosureURL string
	MediaURLs    []string
}

// FeedMetadata contains metadata about an RSS feed.
type FeedMetadata struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Language    string    `json:"language,omitempty"`
	Copyright   string    `json:"copyright,omitempty"`
	Editor      string    `json:"editor,omitempty"`
	WebMaster   string    `json:"webmaster,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	LastBuiltAt time.Time `json:"last_built_at,omitempty"`
	UpdateFreq  int       `json:"update_frequency,omitempty"` // TTL in minutes
	Generator   string    `json:"generator,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	ItemCount   int       `json:"item_count"`
}

// ParsingOptions contains options for RSS parsing.
type ParsingOptions struct {
	// MaxItems limits the number of items to parse (0 = no limit)
	MaxItems int `json:"max_items"`

	// IncludeContent determines whether to include full content
	IncludeContent bool `json:"include_content"`

	// ExtractImages determines whether to extract images from content
	ExtractImages bool `json:"extract_images"`

	// SanitizeHTML determines whether to sanitize HTML content
	SanitizeHTML bool `json:"sanitize_html"`

	// ParseDates determines whether to parse date strings
	ParseDates bool `json:"parse_dates"`

	// FilterDuplicates determines whether to filter duplicate items
	FilterDuplicates bool `json:"filter_duplicates"`

	// MinContentLength filters out items with content shorter than this
	MinContentLength int `json:"min_content_length"`
}

// DefaultParsingOptions returns default parsing options for RSS feeds.
func DefaultParsingOptions() ParsingOptions {
	return ParsingOptions{
		MaxItems:         100,
		IncludeContent:   true,
		ExtractImages:    true,
		SanitizeHTML:     true,
		ParseDates:       true,
		FilterDuplicates: true,
		MinContentLength: 50,
	}
}

// ValidationResult contains the result of RSS feed validation.
type ValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
	ItemCount    int      `json:"item_count"`
	HasContent   bool     `json:"has_content"`
	LastModified string   `json:"last_modified,omitempty"`
}

// FeedStats contains statistics about RSS feed processing.
type FeedStats struct {
	TotalItems     int           `json:"total_items"`
	ValidItems     int           `json:"valid_items"`
	SkippedItems   int           `json:"skipped_items"`
	DuplicateItems int           `json:"duplicate_items"`
	ProcessingTime time.Duration `json:"processing_time"`
	AverageLength  int           `json:"average_content_length"`
	HasImages      int           `json:"items_with_images"`
}

// Constants for RSS parsing
const (
	// MaxFeedSize limits the size of RSS feeds to prevent memory issues
	MaxFeedSize = 10 * 1024 * 1024 // 10 MB

	// MaxItemsPerFeed limits the number of items per feed
	MaxItemsPerFeed = 1000

	// DefaultTimeout for RSS feed fetching
	DefaultTimeout = 30 * time.Second

	// Common RSS date formats
	RFC822  = "Mon, 02 Jan 2006 15:04:05 MST"
	RFC822Z = "Mon, 02 Jan 2006 15:04:05 -0700"
	RFC3339 = "2006-01-02T15:04:05Z07:00"
	ISO8601 = "2006-01-02T15:04:05-07:00"

	// User agent for RSS fetching
	DefaultUserAgent = "NewsAggregator/1.0 (RSS Reader; compatible)"
)
