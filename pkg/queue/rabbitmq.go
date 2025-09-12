package queue

import (
    "encoding/json"
    "fmt"

    "github.com/rs/zerolog/log"
    "github.com/streadway/amqp"

    "news-aggregator/internal/models"
)

type rabbitMQPublisher struct {
    conn     *amqp.Connection
    channel  *amqp.Channel
    exchange string
}

// NewRabbitMQPublisher returns a Publisher backed by RabbitMQ.
func NewRabbitMQPublisher(url string, exchange string) (Publisher, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
    }
    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("failed to open channel: %w", err)
    }
    if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
        ch.Close()
        conn.Close()
        return nil, fmt.Errorf("failed to declare exchange: %w", err)
    }
    return &rabbitMQPublisher{conn: conn, channel: ch, exchange: exchange}, nil
}

func (p *rabbitMQPublisher) Publish(route string, message models.NewsMessage) error {
    body, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }
    return p.channel.Publish(
        p.exchange,
        route,
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        },
    )
}

func (p *rabbitMQPublisher) Close() {
    if p.channel != nil {
        if err := p.channel.Close(); err != nil {
            log.Warn().Err(err).Msg("failed closing rabbitmq channel")
        }
    }
    if p.conn != nil {
        if err := p.conn.Close(); err != nil {
            log.Warn().Err(err).Msg("failed closing rabbitmq connection")
        }
    }
}

type rabbitMQConsumer struct {
    conn     *amqp.Connection
    channel  *amqp.Channel
    exchange string
}

// NewRabbitMQConsumer returns a Consumer backed by RabbitMQ.
func NewRabbitMQConsumer(url string, exchange string, prefetchCount int) (Consumer, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
    }
    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("failed to open channel: %w", err)
    }
    if err := ch.Qos(prefetchCount, 0, false); err != nil {
        ch.Close()
        conn.Close()
        return nil, fmt.Errorf("failed to set qos: %w", err)
    }
    if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
        ch.Close()
        conn.Close()
        return nil, fmt.Errorf("failed to declare exchange: %w", err)
    }
    return &rabbitMQConsumer{conn: conn, channel: ch, exchange: exchange}, nil
}

func (c *rabbitMQConsumer) Consume(queueName string, handler func([]byte) error) error {
    // Declare a queue and bind it to the exchange with routing key matching queueName
    q, err := c.channel.QueueDeclare(queueName, true, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("failed to declare queue: %w", err)
    }
    if err := c.channel.QueueBind(q.Name, queueName, c.exchange, false, nil); err != nil {
        return fmt.Errorf("failed to bind queue: %w", err)
    }
    deliveries, err := c.channel.Consume(q.Name, "", false, false, false, false, nil)
    if err != nil {
        return fmt.Errorf("failed to start consumer: %w", err)
    }
    for d := range deliveries {
        if err := handler(d.Body); err != nil {
            d.Nack(false, true)
            continue
        }
        d.Ack(false)
    }
    return nil
}

func (c *rabbitMQConsumer) Close() {
    if c.channel != nil {
        if err := c.channel.Close(); err != nil {
            log.Warn().Err(err).Msg("failed closing rabbitmq channel")
        }
    }
    if c.conn != nil {
        if err := c.conn.Close(); err != nil {
            log.Warn().Err(err).Msg("failed closing rabbitmq connection")
        }
    }
}


