package main

import (
	"log"

	"main.go/common"
	"main.go/resources"
)

func main() {

	common.ConfigureLogzioSender() // Configure logz.io logger

	// Sending a log message indicating the start of K8S Events Logz.io Integration
	log.Printf("Starting K8S Events Logz.io Integration.")

	// Configuring dynamic client for kubernetes cluster
	common.DynamicClient = common.ConfigureClusterDynamicClient()
	if common.DynamicClient != nil {
		// Adding event handlers if dynamic client is configured successfully
		resources.AddEventHandlers()
	}

	common.LogzioSender.Stop() // Stopping the logz.io logger after the application finishes
}
