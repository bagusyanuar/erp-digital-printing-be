package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/bootstrap"
)

func main() {
	app, err := bootstrap.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Start(); err != nil {
			log.Printf("Server failed to start: %v", err)
			quit <- syscall.SIGTERM // Trigger shutdown on start failure
		}
	}()

	sig := <-quit
	log.Printf("Received signal: %s. Shutting down...", sig.String())
	app.Shutdown()
}
