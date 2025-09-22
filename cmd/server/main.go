package main

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"
	"syscall"
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
	fmt.Println("Connection Successful. Press CTRL+C to exit.")
	<-ctx.Done()
	fmt.Println(" - Signal received. Shutting down.")

}
