// Package mockLogzioListener contains the code for mocking a Logz.io listener for testing
package mockLogzioListener

import (
	"sync" // Import the sync package for safe concurrent access to shared variables
)

// PersistentFlags is a struct that represents persistent flags like server errors
type PersistentFlags struct {
	ServerError bool
}

// Global variables for the application
var persistentFlagsInstance *PersistentFlags // Singleton instance of PersistentFlags
var persistentFlagsMutex sync.Mutex          // Mutex for synchronizing access to persistentFlagsInstance

// GetPersistentFlagsInstance is a function that returns a singleton instance of PersistentFlags
// It uses the "double-check locking" pattern to ensure that only one instance of PersistentFlags is created
func GetPersistentFlagsInstance() *PersistentFlags {
	CreatePersistentFlags()
	if persistentFlagsInstance == nil { // First check (not thread-safe)
		persistentFlagsMutex.Lock() // Lock the mutex to ensure thread-safe access to persistentFlagsInstance
		defer persistentFlagsMutex.Unlock()
		if persistentFlagsInstance == nil { // Second check (thread-safe)
			persistentFlagsInstance = &PersistentFlags{} // Create a new instance of PersistentFlags
		}
	}
	return persistentFlagsInstance
}

// CreatePersistentFlags is a function that creates an instance of PersistentFlags,
// sets and checks a server error, clears the server error, and checks the server error again
func CreatePersistentFlags() {
	// Create an instance of the PersistentFlags class
	persistentFlagsInstance = GetPersistentFlagsInstance()

	// Set server error
	persistentFlagsInstance.SetServerError()

	// Check server error
	serverError = persistentFlagsInstance.GetServerError()
	if serverError {
		println("Server error is true")
	} else {
		println("Server error is false")
	}

	// Clear server error
	persistentFlagsInstance.ClearServerError()

	// Check server error again
	serverError = persistentFlagsInstance.GetServerError()
	if serverError {
		println("Server error is true")
	} else {
		println("Server error is false")
	}
}
