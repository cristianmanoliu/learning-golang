package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// This is a separate entrypoint from main.go.
// Run it with: go run forwarder.go
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	brokers := []string{"localhost:9092"}
	sourceTopic := "demo-topic"
	targetTopic := "demo-topic-forwarded"

	// 👉 make sure target topic exists
	if err := ensureTopic(brokers[0], targetTopic); err != nil {
		log.Fatalf("ensure target topic: %v", err)
	}

	// Reader with a consumer group so offsets are tracked.
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   sourceTopic,
		GroupID: "forwarder-group",
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("close reader: %v\n", err)
		}
	}()

	// Writer to the target topic.
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    targetTopic,
		Balancer: &kafka.LeastBytes{},
	})
	// defer means it will be closed at the end of main()
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("close writer: %v\n", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// cancel is deferred to ensure resources are cleaned up, and it will be called at the end of main()
	defer cancel()

	log.Printf("waiting to fetch one message from topic=%s\n", sourceTopic)

	// FetchMessage gives us manual control over when to commit.
	msg, err := reader.FetchMessage(ctx)
	if err != nil {
		log.Fatalf("fetch message: %v", err)
	}

	log.Printf("read message at offset=%d key=%s value=%s\n",
		msg.Offset, string(msg.Key), string(msg.Value))

	// "Transactional-like" behavior:
	// 1) Write to target topic
	// 2) Only if that succeeded, commit the source offset.
	forwardedValue := fmt.Sprintf("forwarded: %s", string(msg.Value))

	if err := writer.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key,
		Value: []byte(forwardedValue),
	}); err != nil {
		// No commit → on retry, this message will be processed again.
		log.Fatalf("write to target topic failed: %v", err)
	}

	// Only now do we commit the offset in the source topic.
	if err := reader.CommitMessages(ctx, msg); err != nil {
		log.Fatalf("commit message: %v", err)
	}

	log.Printf("successfully forwarded message to topic=%s and committed offset\n", targetTopic)
}

// ensureTopic creates the topic if it doesn't exist.
func ensureTopic(brokerAddr, topic string) error {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		return fmt.Errorf("dial broker: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("get controller: %w", err)
	}

	controllerAddr := fmt.Sprintf("%s:%d", controller.Host, controller.Port)

	ctrlConn, err := kafka.Dial("tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("dial controller: %w", err)
	}
	defer ctrlConn.Close()

	cfgs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	if err := ctrlConn.CreateTopics(cfgs...); err != nil {
		log.Printf("CreateTopics (maybe already exists): %v\n", err)
	}

	return nil
}
