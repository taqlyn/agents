package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/taqlyn/agents/internal/config"
	"github.com/taqlyn/agents/internal/server"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Run(ctx, cfg); err != nil {
		log.Fatal(err)
	}
}
