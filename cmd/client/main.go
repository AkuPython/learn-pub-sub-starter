package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	cs := "amqp://guest:guest@localhost:5672/"
	c, err := amqp.Dial(cs)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	defer c.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	fmt.Println("Connection Successful. Press CTRL+C to exit.")

	name, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	ch, q, err := pubsub.DeclareAndBind(
		c,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, name),
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	defer ch.Close()

	gs := gamelogic.NewGameState(name)

	err = pubsub.SubscribeJSON(c, routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%s", name), routing.PauseKey,
		pubsub.Transient, handlerPause(gs))
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	err = pubsub.SubscribeJSON(c, routing.ExchangePerilTopic,
		fmt.Sprintf("army_moves.%s", name), fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.Transient, handlerMove(gs))
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	quit := false
	go func() {
		<-ctx.Done()
		if !quit {
			fmt.Println(" - Signal received. Shutting down.")
		}
		os.Exit(0)
	}()

InfiniteLoop:
	for {
		uInput := gamelogic.GetInput()
		if len(uInput) == 0 {
			continue
		}
		first := uInput[0]
		switch first {
		case "spawn":
			err := gs.CommandSpawn(uInput)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			move, err := gs.CommandMove(uInput)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, name),
				move)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("move sucessfully published", move.Player.Username)
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			quit = true
			break InfiniteLoop
		default:
			fmt.Println("unknown command")
		}
	}

	fmt.Printf("q: %s --- Msgs: %d --- Cons: %d\n", q.Name, q.Messages, q.Consumers)

}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(gl gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(gl)
		switch moveOutcome {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			return pubsub.Ack
		}
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}
