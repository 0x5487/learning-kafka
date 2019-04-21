package main

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	fmt.Println("== payment start ==")

	topic := "order"

	// receiver from order topic
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost:9092,localhost:9093,localhost:9094",
		"group.id":           "paymentGroup",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	})
	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{topic}, nil)

	handedMessageCount := 0
	for {
		msg, err := c.ReadMessage(-1)
		if err != nil {
			fmt.Printf("[order] Consumer error: %v (%v)\n", err, msg)
			continue
		}

		for _, h := range msg.Headers {
			if h.Key == "action" {
				val := string(h.Value)
				if val == "order-created-event" {
					fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
				} else {
					fmt.Printf("Message on %s: %s\n", msg.TopicPartition, "drop")
				}
			}
		}

		handedMessageCount++
		time.Sleep(1 * time.Second)
		c.Commit()
	}

	c.Close()

	fmt.Println("done")
}
