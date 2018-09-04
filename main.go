package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/3dsinteractive/confluent-kafka-go/kafka"
	"github.com/3dsinteractive/echo"
	randomdata "github.com/Pallinder/go-randomdata"
)

type readOp struct {
	resp chan int
}
type writeOp struct {
	resp chan bool
}

func consume(groupID string, worker string, topic string) {
	workerName := groupID + ":" + worker
	fmt.Printf("%s: Bootstrap Kafka with server %s\n", workerName, os.Getenv("HOST_IP_ADDRESS"))
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("HOST_IP_ADDRESS"),
		"group.id":          groupID,
		"auto.offset.reset": "latest",
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

func produce(producerName string, topic string, writes chan *writeOp) {

	fmt.Printf("%s: Bootstrap Kafka with server %s\n", producerName, os.Getenv("HOST_IP_ADDRESS"))
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("HOST_IP_ADDRESS"),
	})
	if err != nil {
		panic(err)
	}
	defer p.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("%s: Delivery failed: %v\n", producerName, ev.TopicPartition)
				} else {
					fmt.Printf("%s: Delivered message to %v\n", producerName, ev.TopicPartition)
				}
			}
		}
	}()

	// Produce messages to topic (asynchronously)
	for {
		word := randomdata.SillyName()
		fmt.Printf("%s: produce message %s\n", producerName, word)
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}, nil)

		write := &writeOp{
			resp: make(chan bool),
		}
		writes <- write
		<-write.resp

		// time.Sleep(time.Millisecond * 100)
	}
}

func main() {

	writes := make(chan *writeOp)
	reads := make(chan *readOp)

	go func() {
		var counter = 0
		for {
			select {
			case read := <-reads:
				read.resp <- counter
			case write := <-writes:
				counter++
				write.resp <- true
			}
		}
	}()

	go produce("Producer1", "playwithkafka", writes)
	go produce("Producer2", "playwithkafka", writes)
	go produce("Producer3", "playwithkafka", writes)

	go consume("GroupD1", "Worker1", "playwithkafka")
	go consume("GroupD1", "Worker2", "playwithkafka")
	go consume("GroupD2", "Worker1", "playwithkafka")
	go consume("GroupD3", "Worker1", "playwithkafka2")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		read := &readOp{
			resp: make(chan int),
		}
		reads <- read
		counter := <-read.resp

		fmt.Println("Number of messages =", counter)
		os.Exit(1)
	}()

	engine := echo.New()
	engine.GET("/ping", func(c echo.Context) error {
		fmt.Println("Recive ping from", c.Request().Host, c.Request().Header.Get("Host"))
		c.String(http.StatusOK, "pong")
		return nil
	})
	engine.Start(":8080")
}
