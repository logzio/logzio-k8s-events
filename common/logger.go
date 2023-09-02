package common

import (
	"encoding/json"
	"fmt"
	"github.com/logzio/logzio-go"
	"log"
	"os"
	"time"
)

var LogzioLogger *logzio.LogzioSender

func ConfigureLogzioLogger() (LogzioLogger *logzio.LogzioSender) {
	// Creates a resources using Logz.io output configuration: https://app.logz.io/#/dashboard/send-your-data/log-sources/go
	var err error
	LogzioToken := os.Getenv("LOGZIO_TOKEN") // Log shipping token for Logz.io
	if LogzioToken != "" {
		LogzioListener := os.Getenv("LOGZIO_LISTENER")
		if LogzioListener == "" {
			LogzioListener = "https://listener.logz.io:8071" // Defaults to us-east-1 region
		}
		LogzioLogger, err = logzio.New(
			LogzioToken,
			logzio.SetDebug(os.Stderr),
			logzio.SetUrl(LogzioListener),
			logzio.SetDrainDuration(time.Second*5),
			logzio.SetTempDirectory("myQueue"),
			logzio.SetDrainDiskThreshold(99),
		)
		if err != nil {
			log.Fatalf("\n[FATAL] Failed to configure the Logz.io resources.\nERROR: %v\n", err)
		}
	} else {
		log.Fatalf("\n[FATAL] Invalid token configured for LOGZIO_TOKEN environemt variable.\n")
	}
	return LogzioLogger
}
func shipLogMessage(message string) {

	log.Printf("\n[LOG]: %s\n", message)
	err := LogzioLogger.Send([]byte(message))
	if err != nil {
		log.Printf("\nFailed to send log:\n%v to Logz.io.\nRelated error:\n%v.", message, err)
		return
	}

	LogzioLogger.Drain()
}
func SendLog(msg string, extraFields ...interface{}) {
	var err error
	var parsedEventLog []byte
	var logMap map[string]interface{}
	environmentID := os.Getenv("ENV_ID")
	logType := os.Getenv("LOG_TYPE")
	if logType == "" {
		logType = "logzio-informer-events"
	}
	logEvent := LogEvent{Message: msg, Type: logType, EnvironmentID: environmentID}

	if len(extraFields) > 0 {
		extra := fmt.Sprintf("%s", extraFields...)

		log.Printf("\n[DEBUG] Attemping to parse log extra data(%T): %s\tlog(%T):\n%v to Logz.io.\n", extra, extra, logEvent, logEvent)

		if err := json.Unmarshal([]byte(extra), &logEvent); err != nil && extra != "" {
			log.Printf("\n[ERROR] Failed to parse log extra data(%T): %s\tlog(%T):\n%v to Logz.io.\nRelated error:\n%v", extra, logEvent, extra, logEvent, err)

		}
	}
	logByte, _ := json.Marshal(&logEvent)
	json.Unmarshal(logByte, &logMap)
	parsedLogMap := parseLogzioLimits(logMap)
	parsedEventLog, err = json.Marshal(parsedLogMap)
	if err != nil {
		log.Printf("\n[ERROR] Failed to parse event log:\n%v\nERROR:\n%v", logEvent, err)
	}

	message := fmt.Sprintf("%s", string(parsedEventLog))
	if message == "" {
		log.Printf("\n[DEBUG]: Empty message, not sending to Logz.io.\n")
	} else {
		go shipLogMessage(message)
	}

}
