package queue

import "news-aggregator/internal/models"

// Publisher publishes messages to a topic/route
type Publisher interface {
    Publish(route string, message models.NewsMessage) error
    Close()
}

// Consumer consumes messages from a queue/topic
type Consumer interface {
    Consume(queueName string, handler func([]byte) error) error
    Close()
}


