package messaging

import (
	"time"
	"news-aggregator/internal/models/news"
)

// NewsMessage represents a message in the news processing pipeline
type NewsMessage struct {
	ID        string     `json:"id"`
	Source    string     `json:"source"`
	Type      string     `json:"type"` // raw, processed, enriched, indexed
	Data      news.News  `json:"data"`
	Metadata  Metadata   `json:"metadata"`
	Timestamp time.Time  `json:"timestamp"`
	Retry     int        `json:"retry"`
	MaxRetry  int        `json:"max_retry"`
}

// ProcessingResult represents the result of processing a news message
type ProcessingResult struct {
	Success   bool      `json:"success"`
	MessageID string    `json:"message_id"`
	Source    string    `json:"source"`
	Type      string    `json:"type"`
	Error     string    `json:"error,omitempty"`
	Processed news.News `json:"processed,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
}

// Metadata contains additional information about the message
type Metadata struct {
	OriginalURL     string            `json:"original_url"`
	UserAgent       string            `json:"user_agent"`
	CollectionTime  time.Time         `json:"collection_time"`
	ProcessingStage string            `json:"processing_stage"`
	Headers         map[string]string `json:"headers"`
	ContentType     string            `json:"content_type"`
	ContentLength   int64             `json:"content_length"`
	Checksum        string            `json:"checksum"`
}

// DeadLetterMessage represents a message that failed processing
type DeadLetterMessage struct {
	ID           string        `json:"id" db:"id"`
	OriginalMessage NewsMessage `json:"original_message" db:"original_message"`
	FailureReason string       `json:"failure_reason" db:"failure_reason"`
	FailureCount  int          `json:"failure_count" db:"failure_count"`
	LastAttempt   time.Time    `json:"last_attempt" db:"last_attempt"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
}

// QueueStats represents statistics for a message queue
type QueueStats struct {
	QueueName        string        `json:"queue_name"`
	MessageCount     int64         `json:"message_count"`
	ConsumerCount    int           `json:"consumer_count"`
	PublishRate      float64       `json:"publish_rate"`
	DeliveryRate     float64       `json:"delivery_rate"`
	AckRate          float64       `json:"ack_rate"`
	RejectRate       float64       `json:"reject_rate"`
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// ProcessingStats represents processing pipeline statistics
type ProcessingStats struct {
	TotalMessages     int64         `json:"total_messages"`
	ProcessedMessages int64         `json:"processed_messages"`
	FailedMessages    int64         `json:"failed_messages"`
	DeadLetterCount   int64         `json:"dead_letter_count"`
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	ThroughputPerSec  float64       `json:"throughput_per_sec"`
	ErrorRate         float64       `json:"error_rate"`
	LastProcessed     time.Time     `json:"last_processed"`
}

// MessageFilter represents filtering options for messages
type MessageFilter struct {
	Source    string    `json:"source"`
	Type      string    `json:"type"`
	Status    string    `json:"status"` // pending, processed, failed
	DateFrom  time.Time `json:"date_from"`
	DateTo    time.Time `json:"date_to"`
	Page      int       `json:"page"`
	Limit     int       `json:"limit"`
}

// Validation methods

// Validate validates the NewsMessage struct
func (m *NewsMessage) Validate() error {
	if m.ID == "" {
		return ErrEmptyMessageID
	}
	if m.Source == "" {
		return ErrEmptySource
	}
	if m.Type == "" {
		return ErrEmptyMessageType
	}
	
	validTypes := map[string]bool{
		"raw":       true,
		"processed": true,
		"enriched":  true,
		"indexed":   true,
	}
	if !validTypes[m.Type] {
		return ErrInvalidMessageType
	}
	
	if m.MaxRetry < 0 {
		return ErrInvalidMaxRetry
	}
	
	return m.Data.Validate()
}

// Validate validates the MessageFilter struct
func (f *MessageFilter) Validate() error {
	if f.Page < 0 {
		return ErrInvalidPage
	}
	if f.Limit < 0 || f.Limit > 1000 {
		return ErrInvalidLimit
	}
	if !f.DateFrom.IsZero() && !f.DateTo.IsZero() && f.DateFrom.After(f.DateTo) {
		return ErrInvalidDateRange
	}
	
	if f.Type != "" {
		validTypes := map[string]bool{
			"raw":       true,
			"processed": true,
			"enriched":  true,
			"indexed":   true,
		}
		if !validTypes[f.Type] {
			return ErrInvalidMessageType
		}
	}
	
	if f.Status != "" {
		validStatuses := map[string]bool{
			"pending":   true,
			"processed": true,
			"failed":    true,
		}
		if !validStatuses[f.Status] {
			return ErrInvalidStatus
		}
	}
	
	return nil
}

// Helper methods

// ShouldRetry returns true if the message should be retried
func (m *NewsMessage) ShouldRetry() bool {
	return m.Retry < m.MaxRetry
}

// IncrementRetry increments the retry count
func (m *NewsMessage) IncrementRetry() {
	m.Retry++
}

// IsExpired returns true if the message has expired
func (m *NewsMessage) IsExpired(maxAge time.Duration) bool {
	return time.Since(m.Timestamp) > maxAge
}

// GetAge returns the age of the message
func (m *NewsMessage) GetAge() time.Duration {
	return time.Since(m.Timestamp)
}

// SetDefaults sets default values for the MessageFilter
func (f *MessageFilter) SetDefaults() {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = 20
	}
}

// GetOffset returns the offset for pagination
func (f *MessageFilter) GetOffset() int {
	return (f.Page - 1) * f.Limit
}

// NewNewsMessage creates a new news message
func NewNewsMessage(source string, messageType string, data news.News) *NewsMessage {
	return &NewsMessage{
		ID:        generateMessageID(),
		Source:    source,
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
		Retry:     0,
		MaxRetry:  3,
		Metadata: Metadata{
			CollectionTime:  time.Now(),
			ProcessingStage: "initial",
		},
	}
}

// NewProcessingResult creates a new processing result
func NewProcessingResult(messageID, source, msgType string, success bool) *ProcessingResult {
	return &ProcessingResult{
		Success:   success,
		MessageID: messageID,
		Source:    source,
		Type:      msgType,
		Timestamp: time.Now(),
	}
}

// SetError sets an error on the processing result
func (r *ProcessingResult) SetError(err error) {
	r.Success = false
	if err != nil {
		r.Error = err.Error()
	}
}

// SetProcessed sets the processed news and marks as successful
func (r *ProcessingResult) SetProcessed(processedNews news.News) {
	r.Success = true
	r.Processed = processedNews
	r.Error = ""
}

// CalculateErrorRate calculates the error rate for processing stats
func (s *ProcessingStats) CalculateErrorRate() {
	if s.TotalMessages > 0 {
		s.ErrorRate = float64(s.FailedMessages) / float64(s.TotalMessages) * 100
	}
}

// CalculateThroughput calculates the throughput per second
func (s *ProcessingStats) CalculateThroughput(duration time.Duration) {
	if duration.Seconds() > 0 {
		s.ThroughputPerSec = float64(s.ProcessedMessages) / duration.Seconds()
	}
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	// In a real implementation, you might use UUID or a more sophisticated ID generator
	return "msg_" + time.Now().Format("20060102150405") + "_" + randString(8)
}

// randString generates a random string of specified length
func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
