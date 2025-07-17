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
	fmt.Printf("Peril Connection To RabbitMQ Server Successful!\n")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create to State Channel: %v", err)
	}

	gamelogic.PrintServerHelp()

OuterLoop:
	for {
		input := gamelogic.GetInput()

		cmd := input[0]

		switch cmd {
		case "":
			//DO NOTHING
		case "pause":
			fmt.Printf("Pausing Game\n")
			err = pubsub.PublishJSON(publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				})
			if err != nil {
				log.Fatalf("could not publish to State Channel: %v", err)
			}
		case "resume":
			fmt.Printf("Resuming Game\n")
			err = pubsub.PublishJSON(publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				})
			if err != nil {
				log.Fatalf("could not publish to State Channel: %v", err)
			}
		case "quit":
			fmt.Printf("Exiting Game\n")
			break OuterLoop
		default:
			fmt.Print("Not a known command")
		}

	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Printf("RabbitMQ Connection Closed!\n")
}
