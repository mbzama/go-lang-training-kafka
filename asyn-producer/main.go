package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
)

func main() {
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	brokers := []string{"localhost:9092"}

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			panic(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var enqueued, errors int
	doneCh := make(chan struct{})
	go func() {
		for {
			time.Sleep(1000 * time.Millisecond)

			strTime := strconv.Itoa(int(time.Now().Unix()))

			msg := &sarama.ProducerMessage{
				Topic: "test",
				Key:   sarama.StringEncoder(strTime),
				Value: sarama.StringEncoder(fmt.Sprintf("Message-%s", strTime)),
			}
			select {
			case producer.Input() <- msg:
				enqueued++
				fmt.Println(msg)
			case err := <-producer.Errors():
				errors++
				fmt.Println("Failed to produce message:", err)
			case <-signals:
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	log.Printf("Enqueued: %d; errors: %d\n", enqueued, errors)
}
