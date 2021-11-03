package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func console() {
	ctx, ctxCancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(c)
	go func() {
		<-c
		log.Println("Got signal. Canceling the ctx.")
		ctxCancel()
	}()

	T := 10

	log.Printf("Running for %d seconds. Press Ctrl+C to exit sooner.", T)

	for t := T; t > 0; t-- {
		log.Printf("T-%d", t)
		select {
		case <-ctx.Done():
			log.Println("Bye bye")
			os.Exit(123)
			return
		case <-time.After(1 * time.Second):
		}
	}

	log.Println("Bye bye")
}
