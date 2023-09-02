package main

import (
	"main.go/common"
	"main.go/resources"
)

func main() {

	common.LogzioLogger = common.ConfigureLogzioLogger()
	//
	common.SendLog("Starting K8S Events Logz.io Integration.")
	common.DynamicClient = common.ConfigureClusterDynamicClient()
	if common.DynamicClient != nil {
		resources.AddEventHandlers()
	}

	common.LogzioLogger.Stop()
}
