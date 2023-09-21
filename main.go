package main

import (
	"main.go/common"    // Importing common package for application wide functions
	"main.go/resources" // Importing resources package for handling resources
)

func main() {

	common.ConfigureLogzioLogger() // Configure logz.io logger

	// Sending a log message indicating the start of K8S Events Logz.io Integration
	common.SendLog("Starting K8S Events Logz.io Integration.")

	// Configuring dynamic client for kubernetes cluster
	common.DynamicClient = common.ConfigureClusterDynamicClient()
	if common.DynamicClient != nil {
		// Adding event handlers if dynamic client is configured successfully
		resources.AddEventHandlers()
	}

	common.LogzioLogger.Stop() // Stopping the logz.io logger after the application finishes
}
