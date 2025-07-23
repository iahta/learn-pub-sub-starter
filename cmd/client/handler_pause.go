package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(msg routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(msg)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(publishCh,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error: %s", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		}
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, publishCh *amqp.Channel) func(war gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(war gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Printf("> ")
		outcome, winner, loser := gs.HandleWar(war)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			queue := publishGameLog(fmt.Sprintf("%s won a war against %s", winner, loser),
				gs.GetUsername(),
				publishCh)
			return queue
		case gamelogic.WarOutcomeYouWon:
			queue := publishGameLog(fmt.Sprintf("%s won a war against %s", winner, loser),
				gs.GetUsername(),
				publishCh)
			return queue
		case gamelogic.WarOutcomeDraw:
			queue := publishGameLog(fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser),
				gs.GetUsername(),
				publishCh)
			return queue
		default:
			fmt.Printf("error no outcome can be determined")
			return pubsub.NackDiscard
		}
	}
}

func publishGameLog(msg, userName string, publishCh *amqp.Channel) pubsub.Acktype {
	err := pubsub.PublishGob(publishCh,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+userName,
		routing.GameLog{
			CurrentTime: time.Now(),
			Message:     msg,
			Username:    userName,
		})
	if err != nil {
		log.Printf("game log requeue")
		return pubsub.NackRequeue
	}
	log.Printf("game log ack")
	return pubsub.Ack
}
