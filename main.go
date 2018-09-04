package main

import (
	"fmt"
	"net/http"

	"github.com/3dsinteractive/confluent-kafka-go/kafka"
	"github.com/3dsinteractive/echo"
)

func main() {
	go func() {
		c, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": "localhost",
			"group.id":          "myGroup",
			"auto.offset.reset": "earliest",
		})

		if err != nil {
			panic(err)
		}

		c.SubscribeTopics([]string{"playwithkafka"}, nil)

		for {
			msg, err := c.ReadMessage(-1)
			if err == nil {
				fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			} else {
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
				break
			}
		}
		c.Close()
	}()

	engine := echo.New()
	engine.GET("/ping", func(c echo.Context) error {
		c.String(http.StatusOK, "pong")
		return nil
	})
	engine.Start(":8080")
}
