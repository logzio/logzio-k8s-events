// Package mockLogzioListener contains the code for mocking a Logz.io listener for testing
package mockLogzioListener

import "sync" // Import the sync package for safe concurrent access to shared variables

// LogsList is a struct that represents a list of logs
type LogsList struct {
	List []string
}

// Global variables for the application
var logsListInstance *LogsList // Singleton instance of LogsList
var logsListMutex sync.Mutex   // Mutex for synchronizing access to logsListInstance

// GetLogsListInstance is a function that returns a singleton instance of LogsList
// It uses the "double-check locking" pattern to ensure that only one instance of LogsList is created
func GetLogsListInstance() *LogsList {
	if logsListInstance == nil { // First check (not thread-safe)
		logsListMutex.Lock() // Lock the mutex to ensure thread-safe access to logsListInstance
		defer logsListMutex.Unlock()
		if logsListInstance == nil { // Second check (thread-safe)
			logsListInstance = &LogsList{} // Create a new instance of LogsList
		}
	}
	return logsListInstance
}
