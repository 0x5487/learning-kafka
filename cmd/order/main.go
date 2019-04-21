package main

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	uuid "github.com/satori/go.uuid"
)

func main() {
	fmt.Println("== start order service ==")
	topic := "order"

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092,localhost:9093,localhost:9094"})
	if err != nil {
		panic(err)
	}
	defer p.Close()

	// receiver from order topic
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost:9092,localhost:9093,localhost:9094",
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

	// Delivery report handler for produced order message
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
