package mockLogzioListener

import "sync"

type LogsList struct {
	List []string
}

var logsListInstance *LogsList
var logsListMutex sync.Mutex

// GetLogsListInstance is a function that returns a singleton instance of LogsList
// It uses the "double-check locking" pattern to ensure that only one instance of LogsList is created
func GetLogsListInstance() *LogsList {
	if logsListInstance == nil {
		logsListMutex.Lock()
		defer logsListMutex.Unlock()
		if logsListInstance == nil {
			logsListInstance = &LogsList{}
		}
	}
	return logsListInstance
}
