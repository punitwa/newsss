package messaging

import "errors"

// Messaging domain specific errors
var (
	// Validation errors
	ErrEmptyMessageID     = errors.New("message ID cannot be empty")
	ErrEmptySource        = errors.New("message source cannot be empty")
	ErrEmptyMessageType   = errors.New("message type cannot be empty")
	ErrInvalidMessageType = errors.New("invalid message type")
	ErrInvalidMaxRetry    = errors.New("max retry count must be non-negative")
	ErrInvalidPage        = errors.New("page number must be positive")
	ErrInvalidLimit       = errors.New("limit must be between 1 and 1000")
	ErrInvalidDateRange   = errors.New("date from must be before date to")
	ErrInvalidStatus      = errors.New("invalid status")
	
	// Processing errors
	ErrMessageNotFound     = errors.New("message not found")
	ErrMessageExpired      = errors.New("message has expired")
	ErrMaxRetriesExceeded  = errors.New("maximum retries exceeded")
	ErrProcessingFailed    = errors.New("message processing failed")
	ErrSerializationFailed = errors.New("message serialization failed")
	ErrDeserializationFailed = errors.New("message deserialization failed")
	
	// Queue errors
	ErrQueueNotFound      = errors.New("queue not found")
	ErrQueueFull          = errors.New("queue is full")
	ErrQueueEmpty         = errors.New("queue is empty")
	ErrPublishFailed      = errors.New("failed to publish message")
	ErrConsumeFailed      = errors.New("failed to consume message")
	ErrAckFailed          = errors.New("failed to acknowledge message")
	ErrRejectFailed       = errors.New("failed to reject message")
	
	// Connection errors
	ErrConnectionLost     = errors.New("connection to message broker lost")
	ErrConnectionFailed   = errors.New("failed to connect to message broker")
	ErrChannelClosed      = errors.New("message channel is closed")
)
