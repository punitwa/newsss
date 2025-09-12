package search

import (
	"time"
	"news-aggregator/internal/models/news"
)

// Query represents a search query
type Query struct {
	Query      string    `json:"query"`
	Categories []string  `json:"categories"`
	Sources    []string  `json:"sources"`
	Tags       []string  `json:"tags"`
	Authors    []string  `json:"authors"`
	DateFrom   time.Time `json:"date_from"`
	DateTo     time.Time `json:"date_to"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	SortBy     string    `json:"sort_by"`     // relevance, date, popularity
	SortOrder  string    `json:"sort_order"`  // asc, desc
}

// Result represents search results
type Result struct {
	News       []news.News    `json:"news"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Query      string         `json:"query"`
	Took       time.Duration  `json:"took"`
	Facets     *Facets        `json:"facets,omitempty"`
	Suggestions []string      `json:"suggestions,omitempty"`
}

// Facets represents search facets/aggregations
type Facets struct {
	Categories []FacetItem `json:"categories"`
	Sources    []FacetItem `json:"sources"`
	Authors    []FacetItem `json:"authors"`
	Tags       []FacetItem `json:"tags"`
	DateRanges []DateRange `json:"date_ranges"`
}

// FacetItem represents a facet item with count
type FacetItem struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// DateRange represents a date range facet
type DateRange struct {
	Label string    `json:"label"`
	From  time.Time `json:"from"`
	To    time.Time `json:"to"`
	Count int64     `json:"count"`
}

// SavedSearch represents a user's saved search
type SavedSearch struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Query       Query     `json:"query" db:"query"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	Notifications bool    `json:"notifications" db:"notifications"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// SavedSearchRequest represents a request to save a search
type SavedSearchRequest struct {
	Name          string `json:"name" binding:"required"`
	Query         Query  `json:"query" binding:"required"`
	IsDefault     bool   `json:"is_default"`
	Notifications bool   `json:"notifications"`
}

// SearchHistory represents a user's search history
type SearchHistory struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Query     string    `json:"query" db:"query"`
	Filters   Query     `json:"filters" db:"filters"`
	ResultCount int64   `json:"result_count" db:"result_count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Suggestion represents a search suggestion
type Suggestion struct {
	Text  string  `json:"text"`
	Score float64 `json:"score"`
	Type  string  `json:"type"` // query, category, source, author, tag
}

// TrendingQuery represents a trending search query
type TrendingQuery struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
	Trend string `json:"trend"` // up, down, stable
}

// Validation methods

// Validate validates the Query struct
func (q *Query) Validate() error {
	if q.Page < 0 {
		return ErrInvalidPage
	}
	if q.Limit < 0 || q.Limit > 1000 {
		return ErrInvalidLimit
	}
	if !q.DateFrom.IsZero() && !q.DateTo.IsZero() && q.DateFrom.After(q.DateTo) {
		return ErrInvalidDateRange
	}
	
	validSortBy := map[string]bool{
		"":           true,
		"relevance":  true,
		"date":       true,
		"popularity": true,
	}
	if !validSortBy[q.SortBy] {
		return ErrInvalidSortBy
	}
	
	validSortOrder := map[string]bool{
		"":     true,
		"asc":  true,
		"desc": true,
	}
	if !validSortOrder[q.SortOrder] {
		return ErrInvalidSortOrder
	}
	
	return nil
}

// Validate validates the SavedSearch struct
func (s *SavedSearch) Validate() error {
	if s.UserID == "" {
		return ErrEmptyUserID
	}
	if s.Name == "" {
		return ErrEmptySearchName
	}
	return s.Query.Validate()
}

// Validate validates the SavedSearchRequest
func (r *SavedSearchRequest) Validate() error {
	if r.Name == "" {
		return ErrEmptySearchName
	}
	return r.Query.Validate()
}

// Helper methods

// SetDefaults sets default values for the Query
func (q *Query) SetDefaults() {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}
	if q.SortBy == "" {
		q.SortBy = "relevance"
	}
	if q.SortOrder == "" {
		q.SortOrder = "desc"
	}
}

// IsEmpty returns true if the query is empty
func (q *Query) IsEmpty() bool {
	return q.Query == "" && 
		   len(q.Categories) == 0 && 
		   len(q.Sources) == 0 && 
		   len(q.Tags) == 0 && 
		   len(q.Authors) == 0 &&
		   q.DateFrom.IsZero() && 
		   q.DateTo.IsZero()
}

// HasFilters returns true if the query has filters applied
func (q *Query) HasFilters() bool {
	return len(q.Categories) > 0 || 
		   len(q.Sources) > 0 || 
		   len(q.Tags) > 0 || 
		   len(q.Authors) > 0 ||
		   !q.DateFrom.IsZero() || 
		   !q.DateTo.IsZero()
}

// GetOffset returns the offset for pagination
func (q *Query) GetOffset() int {
	return (q.Page - 1) * q.Limit
}

// ToSavedSearch converts SavedSearchRequest to SavedSearch
func (r *SavedSearchRequest) ToSavedSearch(userID string) *SavedSearch {
	return &SavedSearch{
		UserID:        userID,
		Name:          r.Name,
		Query:         r.Query,
		IsDefault:     r.IsDefault,
		Notifications: r.Notifications,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// ToSearchHistory converts Query to SearchHistory
func (q *Query) ToSearchHistory(userID string, resultCount int64) *SearchHistory {
	return &SearchHistory{
		UserID:      userID,
		Query:       q.Query,
		Filters:     *q,
		ResultCount: resultCount,
		CreatedAt:   time.Now(),
	}
}

// HasResults returns true if there are search results
func (r *Result) HasResults() bool {
	return r.Total > 0 && len(r.News) > 0
}

// GetTotalPages returns the total number of pages
func (r *Result) GetTotalPages() int {
	if r.Limit == 0 {
		return 0
	}
	return int((r.Total + int64(r.Limit) - 1) / int64(r.Limit))
}

// HasNextPage returns true if there is a next page
func (r *Result) HasNextPage() bool {
	return r.Page < r.GetTotalPages()
}

// HasPreviousPage returns true if there is a previous page
func (r *Result) HasPreviousPage() bool {
	return r.Page > 1
}

// AddSuggestion adds a suggestion to the result
func (r *Result) AddSuggestion(suggestion string) {
	// Avoid duplicates
	for _, existing := range r.Suggestions {
		if existing == suggestion {
			return
		}
	}
	r.Suggestions = append(r.Suggestions, suggestion)
}
