package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
) error {
	c, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("error declaring queue: %v", err)
	}

	deliveryChan, err := c.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error consuming channel: %v", err)
	}

	go func() {
		for d := range deliveryChan {
			var msg T
			err = json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("error unmarshalling value: %v", err)
			}
			handler(msg)

			err = d.Ack(false)
			if err != nil {
				log.Printf("error acknowledge could not be delivered to the channel: %v", err)
			}
		}
	}()
	return nil
}
