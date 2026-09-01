package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func EnsureTopic(ctx context.Context, brokers []string, topic string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no kafka brokers configured")
	}

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("dial kafka broker: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("get kafka controller: %w", err)
	}

	controllerConn, err := kafka.DialContext(
		ctx,
		"tcp",
		fmt.Sprintf("%s:%d", controller.Host, controller.Port),
	)
	if err != nil {
		return fmt.Errorf("dial kafka controller: %w", err)
	}
	defer controllerConn.Close()

	return controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     3,
		ReplicationFactor: 1,
	})
}

type Producer struct {
	writer *kafka.Writer
	topic  string
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		topic: topic,
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			BatchTimeout: 10 * time.Millisecond,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, key string, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(cfg Config) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			GroupID:        cfg.GroupID,
			Topic:          cfg.Topic,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: 0,
			StartOffset:    kafka.FirstOffset,
		}),
	}
}

func (c *Consumer) Fetch(ctx context.Context) (kafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

func (c *Consumer) Commit(ctx context.Context, message kafka.Message) error {
	return c.reader.CommitMessages(ctx, message)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func ParseBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))
	for _, part := range parts {
		if broker := strings.TrimSpace(part); broker != "" {
			brokers = append(brokers, broker)
		}
	}
	return brokers
}
