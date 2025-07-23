package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Printf("Starting Peril client...\n")
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Printf("Peril game client connected to RabbitMQ!\n")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error creating User Name: %v", err)
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, userName)

	armyKey := fmt.Sprintf("%s.*", routing.ArmyMovesPrefix)
	armyName := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, userName)

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.Transient)
	if err != nil {
		log.Fatalf("error declaring pause: %v", err)
	}

	moveCh, _, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, armyName, armyKey, pubsub.Transient)
	if err != nil {
		log.Fatalf("error declaring move: %v", err)
	}

	gs := gamelogic.NewGameState(userName)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.Transient, handlerPause(gs))
	if err != nil {
		log.Fatalf("could not subsribe to pause: %v", err)
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, armyName, armyKey, pubsub.Transient, handlerMove(gs, moveCh))
	if err != nil {
		log.Fatalf("could not subsribe to army moves: %v", err)
	}

	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(gs, moveCh),
	)
	if err != nil {
		log.Fatalf("could not subscribe to war declaration: %v", err)
	}

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		cmd := input[0]

		switch cmd {
		case "spawn":
			err = gs.CommandSpawn(input)
			if err != nil {
				fmt.Printf("error sending spawn command: %v\n", err)
				continue
			}
		case "move":
			armyMove, err := gs.CommandMove(input)
			if err != nil {
				fmt.Printf("error sending move command: %v\n", err)
				continue
			}
			err = pubsub.PublishJSON(moveCh, routing.ExchangePerilTopic, armyName, armyMove)
			if err != nil {
				fmt.Printf("error publishing move command: %v\n", err)
				continue
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Printf("Spamming not allowed yet!\n")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("command not recognized\n")
			continue
		}
	}

}
