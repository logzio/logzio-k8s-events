package mockLogzioListener

import (
	"sync"
)

type PersistentFlags struct {
	ServerError bool
}

var persistentFlagsInstance *PersistentFlags
var persistentFlagsMutex sync.Mutex

// GetPersistentFlagsInstance is a function that returns a singleton instance of PersistentFlags
// It uses the "double-check locking" pattern to ensure that only one instance of PersistentFlags is created
func GetPersistentFlagsInstance() *PersistentFlags {
	CreatePersistentFlags()
	if persistentFlagsInstance == nil {
		persistentFlagsMutex.Lock()
		defer persistentFlagsMutex.Unlock()
		if persistentFlagsInstance == nil {
			persistentFlagsInstance = &PersistentFlags{}
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
