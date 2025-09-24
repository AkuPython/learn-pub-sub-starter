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
	fmt.Println("Starting Peril server...")
	cs := "amqp://guest:guest@localhost:5672/"
	c, err := amqp.Dial(cs)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	defer c.Close()

	/* This would be just as effective to wait for ctrl+c...
	wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	*/
	// Go Routine compatible version
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	quit := false

	go func() {
		<-ctx.Done()
		if !quit {
			fmt.Println(" - Signal received. Shutting down.")
		}
		os.Exit(0)
	}()

	ch, err := c.Channel()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}

	gamelogic.PrintServerHelp()

InfiniteLoop:
	for {
		uInput := gamelogic.GetInput()
		if len(uInput) == 0 {
			continue
		}
		first := uInput[0]
		switch first {
		case "pause":
			fmt.Println("sending pause message")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect,
				routing.PauseKey, routing.PlayingState{IsPaused: true})
		case "resume":
			fmt.Println("sending resume message")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect,
				routing.PauseKey, routing.PlayingState{IsPaused: false})
		case "quit":
			fmt.Println("exiting game")
			quit = true
			break InfiniteLoop
		default:
			fmt.Println("unknown command")
		}
	}

}
