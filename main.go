package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/3dsinteractive/confluent-kafka-go/kafka"
	"github.com/3dsinteractive/echo"
)

func consume(groupID string, worker string, topic string) {
	workerName := groupID + ":" + worker
	fmt.Printf("%s: Bootstrap Kafka with server %s\n", workerName, os.Getenv("HOST_IP_ADDRESS"))
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("HOST_IP_ADDRESS"),
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	c.Subscribe(topic, nil)
	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			fmt.Printf("%s: Message on %s: %s\n", workerName, msg.TopicPartition, string(msg.Value))
		} else {
			fmt.Printf("%s: Consumer error: %v (%v)\n", workerName, err, msg)
			break
		}
	}
	c.Close()
}

func main() {

	go consume("GroupC1", "Worker1", "playwithkafka")
	go consume("GroupC1", "Worker2", "playwithkafka")
	go consume("GroupC2", "Worker1", "playwithkafka")
	go consume("GroupC3", "Worker1", "playwithkafka2")

	engine := echo.New()
	engine.GET("/ping", func(c echo.Context) error {
		fmt.Println("Recive ping from", c.Request().Host, c.Request().Header.Get("Host"))
		c.String(http.StatusOK, "pong")
		return nil
	})
	engine.Start(":8080")
}
