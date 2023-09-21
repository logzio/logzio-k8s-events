package main

import (
	"log"
	"main.go/common"
	"main.go/mockLogzioListener"
	"testing"
)

func TestDeployEvents(t *testing.T) {
	isListening := mockLogzioListener.SetupMockListener()
	log.Printf("Setup mock listener.")

	if isListening {
		log.Printf("Attempting configuring K8S Events Logz.io logger.")
		common.ConfigureLogzioLogger()

		if common.LogzioLogger != nil {
			common.SendLog("Started K8S Events Logz.io Integration.")
		}

	}

	////
	//
	//common.DynamicClient = common.ConfigureClusterDynamicClient()
	//if common.DynamicClient != nil {
	//	resources.AddEventHandlers()
	//}
	//
	//common.LogzioLogger.Stop()
}
