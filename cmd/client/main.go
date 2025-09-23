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

	fmt.Printf("q: %s --- Msgs: %d --- Cons: %d\n", q.Name, q.Messages, q.Consumers)
	<-ctx.Done()
	fmt.Println(" - Signal received. Shutting down.")

}
