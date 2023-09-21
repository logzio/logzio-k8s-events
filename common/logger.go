package common

import (
	"encoding/json"
	"fmt"
	"github.com/logzio/logzio-go" // Importing logz.io library for logging
	"log"
	"os"
	"sync"
	"time"
)

var LogzioLogger *logzio.LogzioSender // Global variable for logz.io logger
var wg sync.WaitGroup                 // Global variable for wait group

func ConfigureLogzioLogger() {
	// Function to configure logz.io logger
	var err error

	// Reading logz.io token from environment variables
	LogzioToken := os.Getenv("LOGZIO_TOKEN")
	if LogzioToken != "" {
		LogzioListener := os.Getenv("LOGZIO_LISTENER")
		if LogzioListener == "" {
			LogzioListener = "https://listener.logz.io:8071" // Defaults to us-east-1 region
		}
		// Creating a new logz.io logger with specified configuration
		LogzioLogger, err = logzio.New(
			LogzioToken,
			logzio.SetDebug(os.Stderr),
			logzio.SetUrl(LogzioListener),
			logzio.SetDrainDuration(time.Second*5),
			logzio.SetTempDirectory("myQueue"),
			logzio.SetDrainDiskThreshold(99),
		)
		if err != nil {
			// If there is an error in creating the logger, log the error and exit
			log.Fatalf("\n[FATAL] Failed to configure the Logz.io resources.\nERROR: %v\n", err)
		}
	} else {
		// If LOGZIO_TOKEN is not set, log error and exit
		log.Fatalf("\n[FATAL] Invalid token configured for LOGZIO_TOKEN environment variable.\n")
	}
}
func shipLogEvent(eventLog string) {
	// Function to ship log event to logz.io

	// Logging the event
	log.Printf("\n[LOG]: %s\n", eventLog)
	err := LogzioLogger.Send([]byte(eventLog)) // Sending the log event to logz.io
	if err != nil {
		// If there is an error in sending the log, log the error
		log.Printf("\nFailed to send log:\n%v to Logz.io.\nRelated error:\n%v.", eventLog, err)
		return
	}

	LogzioLogger.Drain() // Draining the logger
	defer wg.Done()      // Signaling that this function is done
}
func ParseEventLog(msg string, extraFields ...interface{}) (eventLog string) {
	// This function parses an event log message and any extra fields,
	// converting them into a JSON string.

	var err error
	var parsedEventLog []byte
	var logMap map[string]interface{}

	// Reading environment variables
	environmentID := os.Getenv("ENV_ID")
	logType := os.Getenv("LOG_TYPE")

	if logType == "" {
		logType = "logzio-k8s-events" // Default log type
	}

	// Creating a new log event with the provided message, type and environment ID
	logEvent := LogEvent{Message: msg, Type: logType, EnvironmentID: environmentID}

	if len(extraFields) > 0 {
		// If there are extra fields, convert them to a JSON string and unmarshal into logEvent
		extra := fmt.Sprintf("%s", extraFields...)

		if err = json.Unmarshal([]byte(extra), &logEvent); err != nil && extra != "" {
			// If there is an error in parsing the extra fields, log the error
			log.Printf("\n[ERROR] Failed to parse log extra data(%T): %s\tlog(%T):\n%v to Logz.io.\nRelated error:\n%v", extra, logEvent, extra, logEvent, err)

		}
	}

	// Marshal the log event into a byte slice and unmarshal into logMap
	logByte, _ := json.Marshal(&logEvent)
	json.Unmarshal(logByte, &logMap)

	// Parse the log map to fit logz.io limits
	parsedLogMap := parseLogzioLimits(logMap)

	// Marshal the parsed log map into a byte slice
	parsedEventLog, err = json.Marshal(parsedLogMap)
	if err != nil {
		// If there is an error in marshaling, log the error
		log.Printf("\n[ERROR] Failed to parse event log:\n%v\nERROR:\n%v", logEvent, err)
	}

	// Convert the parsed event log byte slice to a string
	eventLog = fmt.Sprintf("%s", string(parsedEventLog))

	return eventLog
}

func SendLog(msg string, extraFields ...interface{}) {
	// This function sends a log message and any extra fields to logz.io.

	if LogzioLogger != nil {
		// Parse the log message and extra fields into a JSON string
		eventLog := ParseEventLog(msg, extraFields)

		if eventLog == "" {
			// If the parsed event log is empty, drop the log
		} else {
			// Ship the parsed event log to logz.io in a separate goroutine
			go shipLogEvent(eventLog)

			// Increment the wait group counter and wait for all goroutines to finish
			wg.Add(1)
			wg.Wait()
		}
	} else {
		// If the logz.io logger is not configured, log a message and do not send the log
		log.Printf("Logz.io logger isn't configured.\nLog won't be sent:\n%s", msg)
	}

}
