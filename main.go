package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"main.go/common"
	"main.go/resources"
)

func main() {

	common.ConfigureMasking()      // Configure which env var values are masked and how they're hashed
	common.ConfigureLogzioSender() // Configure logz.io logger

	// Sending a log message indicating the start of K8S Events Logz.io Integration
	log.Printf("Starting K8S Events Logz.io Integration.")

	// Kubernetes stops pods with SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	resources.Run(ctx, common.ConfigureClusterDynamicClient())

	log.Printf("Stopping K8S Events Logz.io Integration, sending the remaining logs.")
	common.LogzioSender.Stop() // Stopping the logz.io logger after the application finishes
}
