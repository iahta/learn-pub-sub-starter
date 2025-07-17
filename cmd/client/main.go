package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Printf("Starting Peril client...\n")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error creating User Name: %v", err)
	}

	queueName := routing.PauseKey + "." + userName

	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.Transient)
	if err != nil {
		log.Fatalf("error declaring: %v", err)
	}

	gs := gamelogic.NewGameState(userName)

OuterLoop:
	for {
		input := gamelogic.GetInput()
		cmd := input[0]

		switch cmd {
		case "spawn":
			err = gs.CommandSpawn(input)
			if err != nil {
				log.Fatalf("error sending spawn command: %v", err)
			}
		case "move":
			_, err := gs.CommandMove(input)
			if err != nil {
				log.Fatalf("error sending move command: %v", err)
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Printf("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break OuterLoop
		default:
			fmt.Printf("command not recognized")
			continue
		}

	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Printf("\nRabbitMQ Connection Closed!\n")
}
