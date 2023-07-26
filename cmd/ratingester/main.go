package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"moviedata.com/rating/pkg/model"
)

func main() {
	log.Println("Creating a Kafka producer")
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.server": "localhost"})
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	const filename = "ratingsdata.json"

	fmt.Println("Reading rating events from file" + filename)
	ratingEvents, err := readRatingEvents(filename)
	if err != nil {
		panic(err)
	}
	const topic = "ratings"
	if err := prodceRatingEvents(topic, producer, ratingEvents); err != nil {
		panic(err)
	}
	const timeout = 10 * time.Second
	fmt.Println("Waiting " + timeout.String() + " until all events get produced")
	producer.Flush(int(timeout.Milliseconds()))
}

func readRatingEvents(filename string) ([]model.RatingEvent, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var ratings []model.RatingEvent
	if err := json.NewDecoder(f).Decode(&ratings); err != nil {
		return nil, err
	}
	return ratings, nil
}

func prodceRatingEvents(topic string, producer *kafka.Producer, events []model.RatingEvent) error {
	for _, ratingEvent := range events {
		encodedEvent, err := json.Marshal(ratingEvent)
		if err != nil {
			return err
		}
		if err := producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(encodedEvent),
		}, nil); err != nil {
			return err
		}
	}
	return nil
}
