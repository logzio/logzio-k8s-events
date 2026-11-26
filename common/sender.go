package common

import (
	"encoding/json"
	"fmt"
	"github.com/logzio/logzio-go"
	"log"
	"os"
	"strconv"
	"time"
)

var LogzioSender *logzio.LogzioSender

// debugLogs prints each full event to stdout. Off by default: the pod's own logs are often shipped to Logz.io too,
// which would ingest every event twice
var debugLogs bool

// ConfigureLogzioSender configures the Logz.io sender
func ConfigureLogzioSender() {

	if debug := os.Getenv("DEBUG"); debug != "" {
		var err error
		if debugLogs, err = strconv.ParseBool(debug); err != nil {
			log.Printf("[WARN] Ignoring invalid DEBUG value %q, expected true or false.", debug)
		}
	}

	// Reading logz.io token from environment variables
	LogzioToken := os.Getenv("LOGZIO_TOKEN")
	if LogzioToken != "" {
		LogzioListener := os.Getenv("LOGZIO_LISTENER")
		if LogzioListener == "" {
			LogzioListener = DefaultListener // Defaults to us-east-1 region
		}
		// Creating a new logz.io logger with specified configuration
		var err error
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
func shipLogEvent(msg string, eventLog string) {
	if line := stdoutLine(msg, eventLog); line != "" {
		log.Print(line)
	}
	err := LogzioSender.Send([]byte(eventLog)) // Sending the log event to logz.io
	if err != nil {
		// If there is an error in sending the log, log the error
		log.Printf("\nFailed to send log: %s to Logz.io.\nRelated error:\n%v.", msg, err)
	}
}

// stdoutLine returns what to print for an event: its message, or the full event in debug mode
func stdoutLine(msg string, eventLog string) string {
	if debugLogs {
		return fmt.Sprintf("\n[LOG]: %s\n", eventLog)
	}
	return msg
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
		log.Printf("\n[ERROR] Failed to parse event log: %s\nERROR:\n%v", msg, err)
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
		log.Printf("\n[ERROR] Failed to parse event log: %s\nERROR:\n%v", msg, err)
	}

	return string(parsedEventLog)
}

// SendLog sends a log message and any extra fields to Logz.io.
func SendLog(msg string, extraFields ...map[string]interface{}) {

	if LogzioSender != nil {
		var eventLog string
		if len(extraFields) > 0 {
			eventLog = ParseEventLog(msg, extraFields[0])
		} else {
			eventLog = ParseEventLog(msg)
		}
		if eventLog == "" {
			return
		}
		shipLogEvent(msg, eventLog)
	}
}
