package common

import (
	"encoding/json"
	"github.com/logzio/logzio-go"
	"log"
	"main.go/mockLogzioListener"
	"testing"

	"os"
	"time"
)

// TestStartMockLogzioListener is used to test the mock listener functionality.
func TestStartMockLogzioListener(t *testing.T) {
	mockLogzioListener.StartMockLogzioListener()
	mockListener := mockLogzioListener.MockListener
	if mockListener != nil {
		mockListenerURL := mockLogzioListener.GetMockListenerURL()
		logsCount := mockListener.NumberOfLogs()
		if mockListenerURL != "" {
			err := os.Setenv("LOGZIO_TOKEN", "test-shipping-token")
			if err != nil {
				return
			}
			err = os.Setenv("LOGZIO_LISTENER", mockListenerURL)
			if err != nil {
				return
			}

			if logsCount > 0 {
				t.Log("Successfully sent logs.")
			}
		}
	} else {
		t.Error("Failed to start mock listener")
	}

}

// TestConfigureLogzioSender is used to test the Logz.io sender configuration functionality.
func TestConfigureLogzioSender(t *testing.T) {
	LogzioToken := os.Getenv("LOGZIO_TOKEN") // Log shipping token for Logz.io
	if LogzioToken != "" {
		LogzioListener := os.Getenv("LOGZIO_LISTENER")
		if LogzioListener == "" {
			LogzioListener = DefaultListener // Defaults to us-east-1 region
		}
		LogzioSender, err = logzio.New(
			LogzioToken,
			logzio.SetDebug(os.Stderr),
			logzio.SetUrl(LogzioListener),
			logzio.SetDrainDuration(time.Second*5),
			logzio.SetDrainDiskThreshold(99),
		)
		if err != nil {
			t.Errorf("\n[FATAL] Failed to configure the Logz.io logger.\nERROR: %v\n", err)
		} else {
			t.Log("Successfully configured the Logz.io logger.\n")
		}
	} else {
		t.Error("\n[FATAL] Invalid token configured for LOGZIO_TOKEN environment variable.\n")
	}

}

// TestSendLog is used to test the Logz.io sender functionality.
func TestSendLog(t *testing.T) {

	t.Run("SendLog", func(t *testing.T) {
		os.Setenv("ENV_ID", "dev")
		os.Setenv("LOG_TYPE", "logzio-k8s-events-test")
		logsListInstance := mockLogzioListener.GetLogsListInstance()
		allLogs := logsListInstance.List
		eventLog := GetTestEventLog()
		parsedEventLog, err := json.Marshal(eventLog)
		if err != nil {
			log.Printf("EventLog JSON marshaling failed: %s", err)
		}
		allLogs = append(allLogs, string(parsedEventLog)) // append the log to the list
		for _, _ = range allLogs {
			SendLog("Test log", eventLog)
		}
	})
}
