package news

import "time"

// News represents a news article
type News struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	Summary     string    `json:"summary" db:"summary"`
	URL         string    `json:"url" db:"url"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	Author      string    `json:"author" db:"author"`
	Source      string    `json:"source" db:"source"`
	Category    string    `json:"category" db:"category"`
	Tags        []string  `json:"tags" db:"tags"`
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Hash        string    `json:"-" db:"content_hash"` // For deduplication
}

// Category represents a news category
type Category struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Color       string `json:"color" db:"color"`
	Icon        string `json:"icon" db:"icon"`
}

// Filter represents filtering options for news queries
type Filter struct {
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
	Category string    `json:"category"`
	Source   string    `json:"source"`
	Tags     []string  `json:"tags"`
	DateFrom time.Time `json:"date_from"`
	DateTo   time.Time `json:"date_to"`
}

// Stats contains news-related statistics
type Stats struct {
	TotalArticles     int64           `json:"total_articles"`
	ArticlesToday     int64           `json:"articles_today"`
	ArticlesThisWeek  int64           `json:"articles_this_week"`
	ArticlesThisMonth int64           `json:"articles_this_month"`
	TopCategories     []CategoryStats `json:"top_categories"`
	TopSources        []SourceStats   `json:"top_sources"`
}

// CategoryStats represents statistics for a specific category
type CategoryStats struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// SourceStats represents statistics for a specific source
type SourceStats struct {
	Source string `json:"source"`
	Count  int64  `json:"count"`
}

// Validation methods

// Validate validates the News struct
func (n *News) Validate() error {
	if n.Title == "" {
		return ErrEmptyTitle
	}
	if n.URL == "" {
		return ErrEmptyURL
	}
	if n.Source == "" {
		return ErrEmptySource
	}
	return nil
}

// Validate validates the Category struct
func (c *Category) Validate() error {
	if c.Name == "" {
		return ErrEmptyCategoryName
	}
	return nil
}

// Validate validates the Filter struct
func (f *Filter) Validate() error {
	if f.Page < 0 {
		return ErrInvalidPage
	}
	if f.Limit < 0 || f.Limit > 1000 {
		return ErrInvalidLimit
	}
	if !f.DateFrom.IsZero() && !f.DateTo.IsZero() && f.DateFrom.After(f.DateTo) {
		return ErrInvalidDateRange
	}
	return nil
}

// Helper methods

// IsRecent returns true if the news article was published within the last 24 hours
func (n *News) IsRecent() bool {
	return time.Since(n.PublishedAt) <= 24*time.Hour
}

// HasImage returns true if the news article has an associated image
func (n *News) HasImage() bool {
	return n.ImageURL != ""
}

// GetAge returns the age of the news article
func (n *News) GetAge() time.Duration {
	return time.Since(n.PublishedAt)
}

// SetDefaults sets default values for the Filter
func (f *Filter) SetDefaults() {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = 20
	}
}
