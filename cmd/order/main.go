package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	uuid "github.com/satori/go.uuid"
)

func main() {
	topic := "order"
	brokers := "localhost:9092,localhost:9093,localhost:9094"

	fmt.Println("== start order service ==")

	// receiver from order topic
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":           "orderGroup",
		"enable.auto.commit": "false",
		"auto.offset.reset":  "earliest",
	})
	if err != nil {
		panic(err)
	}

	go func() {
		c.SubscribeTopics([]string{topic}, nil)

		for {
			msg, err := c.ReadMessage(-1)
			if err != nil {
				fmt.Printf("[order] Consumer error: %v (%v)\n", err, msg)
				continue
			}

			for _, h := range msg.Headers {
				if h.Key == "action" {
					val := string(h.Value)
					if val == "order-purchased-event" {
						fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
					} else {
						fmt.Println("[order] drop msg")
					}
				}
			}

			time.Sleep(1 * time.Second)
			c.Commit()
		}

		c.Close()
	}()

	// create topics
	a, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		os.Exit(1)
	}
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		panic("ParseDuration(60s)")
	}
	results, err := a.CreateTopics(
		context.Background(),
		// Multiple topics can be created simultaneously
		// by providing more TopicSpecification structs here.
		[]kafka.TopicSpecification{
			{
				Topic:             topic,
				NumPartitions:     3,
				ReplicationFactor: 3,
				Config:            map[string]string{"retention.ms": "604800000"},
			},
		},
		// Admin options
		kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to create topic: %v\n", err)
		os.Exit(1)
	}
	// Print results
	for _, result := range results {
		fmt.Printf("[create]topic: %s, err: %s\n", result.Topic, result.Error)
		//kafkaError, ok := result.Error.(kafka.Error)
		if result.Error.Code() == kafka.ErrTopicAlreadyExists {
			fmt.Println("topic order already exists")
		}
		if result.Error.Code() == kafka.ErrNoError {
			fmt.Println("create topic success")
		}
	}

	// Delivery report handler for produced order message
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		panic(err)
	}
	defer p.Close()

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Produce messages to topic (asynchronously)
	i := 0
	for {
		orderID := uuid.NewV4()
		payload := fmt.Sprintf("no: %d, orderID: %s purchased", i, orderID)
		fmt.Println(payload)

		header := kafka.Header{
			Key:   "action",
			Value: []byte("order-purchased-event"),
		}
		headers := []kafka.Header{}
		headers = append(headers, header)

		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value:   []byte(payload),
			Key:     []byte(orderID.String()),
			Headers: headers,
		}

		p.Produce(msg, nil)
		i++
		time.Sleep(2 * time.Second)
	}

	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)

	fmt.Println("== done order service ==")
}
