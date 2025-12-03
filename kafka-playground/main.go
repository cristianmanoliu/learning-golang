package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	brokers := []string{"localhost:9092"}
	topic := "demo-topic"

	// 👉 Make sure topic exists
	if err := ensureTopic(brokers[0], topic); err != nil {
		log.Fatalf("ensure topic: %v", err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	// defer means "execute this when the surrounding function returns"
	// surrounding function here is main()
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("close writer: %v\n", err)
		}
	}()

	msgValue := fmt.Sprintf("hello from Go at %s", time.Now().Format(time.RFC3339))
	log.Printf("publishing message to topic=%s: %s\n", topic, msgValue)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer means "execute this when the surrounding function returns"
	// surrounding function here is main()
	defer cancel()

	if err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("key-1"),
		Value: []byte(msgValue),
	}); err != nil {
		log.Fatalf("failed to write message: %v", err)
	}

	log.Println("message published successfully")
}

// ensureTopic creates the topic if it doesn't exist.
func ensureTopic(brokerAddr, topic string) error {
	// Connect to the first broker
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		return fmt.Errorf("dial broker: %w", err)
	}
	defer conn.Close()

	// Ask cluster who the controller is
	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("get controller: %w", err)
	}

	// controller address is used to create topics
	controllerAddr := fmt.Sprintf("%s:%d", controller.Host, controller.Port)

	// Connect to the controller
	ctrlConn, err := kafka.Dial("tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("dial controller: %w", err)
	}
	defer ctrlConn.Close()

	// Create topic with 1 partition, replication factor 1
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	if err := ctrlConn.CreateTopics(topicConfigs...); err != nil {
		// Kafka will also return an error if topic already exists; for a playground we can just log it.
		log.Printf("CreateTopics (maybe already exists): %v\n", err)
	}

	// return nil means no error
	return nil
}
