package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go_chat_bot/pkg/chat"
)

func main() {
	var port int
	var logfile string
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.StringVar(&logfile, "logfile", "", "optional logfile path")
	flag.Parse()

	if logfile != "" {
		f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to open logfile: %v", err)
		}
		log.SetOutput(f)
		defer f.Close()
	}

	svc := chat.NewChatService()
	if err := svc.RegisterService(); err != nil {
		log.Fatalf("failed to register RPC service: %v", err)
	}

	// Create stop channel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("START server port=%d", port)

	// Start RPC server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- svc.ServeRPC("tcp", addr)
	}()

	// Wait for either server error or shutdown signal
	select {
	case err := <-errChan:
		if err != nil {
			log.Printf("server error: %v", err)
		}
	case sig := <-stop:
		log.Printf("received signal %v, shutting down", sig)
		svc.Shutdown() // This will close the listener and make ServeRPC return
	}

	log.Printf("SHUTDOWN complete")
}
