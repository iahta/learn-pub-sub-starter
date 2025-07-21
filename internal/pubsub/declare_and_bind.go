package pubsub

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

func DeclareAndBind(conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	declareCh, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create to declare channel: %v", err)
	}

	var newQueue amqp.Queue

	deadTable := amqp.Table{
		"x-dead-letter-exchange": routing.ExchangePerilDead,
	}

	switch queueType {
	case Durable:
		newQueue, err = declareCh.QueueDeclare(queueName, true, false, false, false, deadTable)
		if err != nil {
			declareCh.Close()
			return nil, amqp.Queue{}, fmt.Errorf("error declaring durable queue: %v", err)
		}
	case Transient:
		newQueue, err = declareCh.QueueDeclare(queueName, false, true, true, false, deadTable)
		if err != nil {
			declareCh.Close()
			return nil, amqp.Queue{}, fmt.Errorf("error declaring transient queue: %v", err)
		}
	default:
		declareCh.Close()
		return nil, amqp.Queue{}, fmt.Errorf("unsupported queue type: %v", queueType)
	}
	err = declareCh.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		declareCh.Close()
		return nil, amqp.Queue{}, fmt.Errorf("error binding queue: %v", err)
	}
	return declareCh, newQueue, nil
}
