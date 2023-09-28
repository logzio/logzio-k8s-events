package common

import (
	"encoding/json"
	"fmt"
	"github.com/logzio/logzio-go"
	"log"
	"os"
	"sync"
	"time"
)

var LogzioSender *logzio.LogzioSender
var wg sync.WaitGroup

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
			logzio.SetDebug(os.Stderr),
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
	defer wg.Done()      // Signaling that this function is done
}

// ParseEventLog function parses an event log message and any extra fields,
// converting them into a JSON string.
func ParseEventLog(msg string, extraFields ...interface{}) (eventLog string) {

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
		// If there are extra fields, convert them to a JSON string and unmarshal into logEvent
		extra := fmt.Sprintf("%s", extraFields...)

		if err = json.Unmarshal([]byte(extra), &logEvent); err != nil && extra != "" && extra != "[]" {
			// If there is an error in parsing the extra fields, log the error
			log.Printf("\n[ERROR] Failed to parse log extra data(%T): %s\tlog(%T):\n%v to Logz.io.\nRelated error:\n%v", extra, extra, logEvent, logEvent, err)

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

	return string(parsedEventLog)
}

// SendLog sends a log message and any extra fields to Logz.io.
func SendLog(msg string, extraFields ...interface{}) {

	if LogzioSender != nil {
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
	}
}
