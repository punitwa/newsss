package user

import (
	"time"
	"news-aggregator/internal/models/news"
)

// Bookmark represents a user's bookmarked article
type Bookmark struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	NewsID    string     `json:"news_id" db:"news_id"`
	News      *news.News `json:"news,omitempty"`
	Tags      []string   `json:"tags" db:"tags"`           // User-defined tags for the bookmark
	Notes     string     `json:"notes" db:"notes"`         // User notes about the bookmark
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// BookmarkRequest represents a request to bookmark an article
type BookmarkRequest struct {
	NewsID string   `json:"news_id" binding:"required"`
	Tags   []string `json:"tags,omitempty"`
	Notes  string   `json:"notes,omitempty"`
}

// UpdateBookmarkRequest represents a request to update a bookmark
type UpdateBookmarkRequest struct {
	Tags  []string `json:"tags,omitempty"`
	Notes string   `json:"notes,omitempty"`
}

// BookmarkFilter represents filtering options for bookmarks
type BookmarkFilter struct {
	UserID   string    `json:"user_id"`
	Tags     []string  `json:"tags"`
	DateFrom time.Time `json:"date_from"`
	DateTo   time.Time `json:"date_to"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

// Validation methods

// Validate validates the Bookmark struct
func (b *Bookmark) Validate() error {
	if b.UserID == "" {
		return ErrEmptyUserID
	}
	if b.NewsID == "" {
		return ErrEmptyNewsID
	}
	return nil
}

// Validate validates the BookmarkRequest
func (r *BookmarkRequest) Validate() error {
	if r.NewsID == "" {
		return ErrEmptyNewsID
	}
	return nil
}

// Validate validates the BookmarkFilter
func (f *BookmarkFilter) Validate() error {
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

// HasTag checks if the bookmark has a specific tag
func (b *Bookmark) HasTag(tag string) bool {
	for _, t := range b.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag adds a tag to the bookmark if it doesn't already exist
func (b *Bookmark) AddTag(tag string) {
	if !b.HasTag(tag) {
		b.Tags = append(b.Tags, tag)
		b.UpdatedAt = time.Now()
	}
}

// RemoveTag removes a tag from the bookmark
func (b *Bookmark) RemoveTag(tag string) {
	for i, t := range b.Tags {
		if t == tag {
			b.Tags = append(b.Tags[:i], b.Tags[i+1:]...)
			b.UpdatedAt = time.Now()
			break
		}
	}
}

// SetDefaults sets default values for the BookmarkFilter
func (f *BookmarkFilter) SetDefaults() {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = 20
	}
}

// ToBookmark converts BookmarkRequest to Bookmark
func (r *BookmarkRequest) ToBookmark(userID string) *Bookmark {
	return &Bookmark{
		UserID:    userID,
		NewsID:    r.NewsID,
		Tags:      r.Tags,
		Notes:     r.Notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ApplyToBookmark applies the update request to a bookmark
func (r *UpdateBookmarkRequest) ApplyToBookmark(bookmark *Bookmark) {
	if r.Tags != nil {
		bookmark.Tags = r.Tags
	}
	if r.Notes != "" {
		bookmark.Notes = r.Notes
	}
	bookmark.UpdatedAt = time.Now()
}
