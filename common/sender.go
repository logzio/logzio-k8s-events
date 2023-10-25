package common

import (
	"encoding/json"
	"github.com/logzio/logzio-go"
	"log"
	"os"
	"sync"
	"time"
)

var LogzioSender *logzio.LogzioSender

// ConfigureLogzioSender configures the Logz.io sender
func ConfigureLogzioSender() {

	// Reading logz.io token from environment variables
	LogzioToken := os.Getenv("LOGZIO_TOKEN")
	if LogzioToken != "" {
		LogzioListener := os.Getenv("LOGZIO_LISTENER")
		if LogzioListener == "" {
			LogzioListener = DefaultListener // Defaults to us-east-1 region
		}
		// Creating a new logz.io logger with specified configuration
		LogzioSender, err = logzio.New(
			LogzioToken,
			logzio.SetUrl(LogzioListener),
			logzio.SetDrainDuration(time.Second*5),
			logzio.SetDrainDiskThreshold(99),
		)
		if err != nil {
			// If there is an error in creating the logger, log the error and exit
			log.Fatalf("\n[FATAL] Failed to configure the Logz.io resources.\nERROR: %v\n", err)
		}
	} else {
		// If LOGZIO_TOKEN is not set, log error and exit
		log.Fatalf("\n[FATAL] Invalid shipping token configured for LOGZIO_TOKEN environment variable.\n")
	}
}

// shipLogEvent ships log event to Logz.io
func shipLogEvent(eventLog string) {

	// Logging the event
	log.Printf("\n[LOG]: %s\n", eventLog)
	err = LogzioSender.Send([]byte(eventLog)) // Sending the log event to logz.io
	if err != nil {
		// If there is an error in sending the log, log the error
		log.Printf("\nFailed to send log:\n%v to Logz.io.\nRelated error:\n%v.", eventLog, err)
		return
	}

	LogzioSender.Drain() // Draining the logger

}

func ParseEventLog(msg string, extraFields ...map[string]interface{}) (eventLog string) {

	var parsedEventLog []byte
	var logMap map[string]interface{}

	// Reading environment variables
	environmentID := os.Getenv("ENV_ID")
	logType := os.Getenv("LOG_TYPE")

	if logType == "" {
		logType = DefaultLogType // Default log type
	}

	// Creating a new log event with the provided message, type and environment ID
	logEvent := LogEvent{Message: msg, Type: logType, EnvironmentID: environmentID}

	if len(extraFields) > 0 {
		logEvent.ExtraFields = extraFields[0]
	}

	// Marshal the log event into a byte slice and unmarshal into logMap
	logByte, _ := json.Marshal(&logEvent)

	err := json.Unmarshal(logByte, &logMap)
	if err != nil {
		// If there is an error in unmarshaling, log the error
		log.Printf("\n[ERROR] Failed to parse event log:\n%v\nERROR:\n%v", logEvent, err)
		return
	}

	// Merge ExtraFields into logMap
	for key, value := range logEvent.ExtraFields {
		logMap[key] = value
	}
	// Parse the log map to fit logz.io limits
	parsedLogMap := parseLogzioLimits(logMap)

	// Marshal the parsed log map into a byte slice
	parsedEventLog, err = json.Marshal(parsedLogMap)
	if err != nil {
		// If there is an error in marshaling, log the error
		log.Printf("\n[ERROR] Failed to parse event log:\n%v\nERROR:\n%v", logEvent, err)
	}

	return string(parsedEventLog)
}

// SendLog sends a log message and any extra fields to Logz.io.
func SendLog(msg string, extraFields ...map[string]interface{}) {

	if LogzioSender != nil {
		var senderWG sync.WaitGroup // Create a new WaitGroup here
		var eventLog string
		if len(extraFields) > 0 {
			eventLog = ParseEventLog(msg, extraFields[0])
		} else {
			eventLog = ParseEventLog(msg)
		}
		if eventLog == "" {
			return
		}

		senderWG.Add(1) // Increment the WaitGroup
		go func() {
			defer senderWG.Done() // Decrement the WaitGroup when the goroutine finishes
			shipLogEvent(eventLog)

		}()

		senderWG.Wait() // Wait for all goroutines to finish
	}
}
