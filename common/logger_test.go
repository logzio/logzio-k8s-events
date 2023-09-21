package common

import (
	"github.com/logzio/logzio-go"
	"main.go/mockLogzioListener"
	"testing"

	"os"
	"time"
)

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

func TestConfigureLogzioLogger(t *testing.T) {
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
			t.Errorf("\n[FATAL] Failed to configure the Logz.io logger.\nERROR: %v\n", err)
		} else {
			t.Log("Successfully configured the Logz.io logger.\n")
		}
	} else {
		t.Error("\n[FATAL] Invalid token configured for LOGZIO_TOKEN environment variable.\n")
	}

}
func TestSendLog(t *testing.T) {

	t.Run("SendLog", func(t *testing.T) {
		os.Setenv("ENV_ID", "dev")
		os.Setenv("LOG_TYPE", "logzio-k8s-events-test")
		logsListInstance := mockLogzioListener.GetLogsListInstance()
		allLogs := logsListInstance.List
		for _, testLog := range allLogs {
			SendLog("Test log", testLog)
		}
	})
}
