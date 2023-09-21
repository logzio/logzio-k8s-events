package mockLogzioListener

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

// ListenerHandler is an empty struct that will implement the http.Handler interface
type ListenerHandler struct{}

// MockLogzioListener is a struct that represents a mock Logz.io listener for testing
type MockLogzioListener struct {
	Port            int
	Host            string
	LogsList        []string
	PersistentFlags *PersistentFlags
	server          *http.Server
	listeningThread *Thread
}

// Global variables for the application
var logsList LogsList
var persistentFlags *PersistentFlags
var mutex sync.Mutex                 // Mutex for synchronizing access to shared variables
var serverError bool                 // Flag to simulate server errors
var MockListener *MockLogzioListener // The mock Logz.io listener

// ServeHTTP is the function that gets called on each HTTP request
func (h *ListenerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" { // Handle POST requests
		logsList = *GetLogsListInstance()

		// Define the structure of the expected request body
		type RequestBody struct {
			Message string `json:"message"`
		}

		// Read the request body
		reqBody, err := io.ReadAll(r.Body)

		var requestBody RequestBody
		json.Unmarshal(reqBody, &requestBody)
		b, err := json.Marshal(requestBody)

		fmt.Printf("%s", b)
		if err != nil {
			http.Error(w, fmt.Sprintf("Bad Request\nRequest:\n%v", requestBody), http.StatusBadRequest)
			return
		}

		// Log the received POST request
		log.Printf("Received POST request to: %s \nRequest:\n%v", r.Host, requestBody)

		// Split the logs by new line
		allLogs := strings.Split(string(b), "\n")

		// If no logs are received, return a bad request error
		if len(allLogs) == 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		mutex.Lock() // Lock the mutex to ensure thread-safe access to shared variables
		defer mutex.Unlock()

		// If the serverError flag is set, return an internal server error
		if serverError {
			http.Error(w, "Issue!!!!!!!", http.StatusInternalServerError)
			return
		}

		// Append each log to the logs list
		for _, testLog := range allLogs {
			if testLog != "" {
				logsList.List = append(logsList.List, testLog)
			}
		}

		// Return a success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Shabam! got logs."))
	} else { // For all other HTTP methods, return a method not allowed error
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// NewMockLogzioListener function creates a new instance of MockLogzioListener
// It finds an available port and initializes the server with that port and a listener handler
func NewMockLogzioListener() *MockLogzioListener {
	// find an available port
	port, err := findAvailablePort()
	if err != nil {
		fmt.Println("Error finding available port:", err)
	}

	// define the host
	host := "localhost"

	// create a new listener handler
	listenerHandler := &ListenerHandler{}

	// create a new http server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: listenerHandler,
	}

	// create a new thread for listening
	listeningThread := &Thread{
		target: startListening,
	}

	// return a new instance of MockLogzioListener
	return &MockLogzioListener{
		Port:            port,
		Host:            host,
		LogsList:        logsList.List,
		PersistentFlags: persistentFlags,
		server:          server,
		listeningThread: listeningThread,
	}
}

// StartListening is a method that starts the listening thread
func (ml *MockLogzioListener) StartListening() {
	ml.listeningThread.Start()
}

// startListening is a function that starts the http server
func startListening() {
	err := MockListener.server.ListenAndServe()
	if err != nil {
		fmt.Println("HTTP server error:", err)
	}
}

// FindLog is a method that locks the mutex and searches the logs list
// for a specific log string, returning a boolean value
func (ml *MockLogzioListener) FindLog(searchLog string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	// search for the log in the logs list
	for _, currentLog := range ml.LogsList {
		if strings.Contains(currentLog, searchLog) {
			return true
		}
	}

	return false
}

// NumberOfLogs is a method that locks the mutex and returns the number of logs in the logs list
func (ml *MockLogzioListener) NumberOfLogs() int {
	mutex.Lock()
	defer mutex.Unlock()

	// return the number of logs
	return len(ml.LogsList)
}

// ClearLogsBuffer is a method that locks the mutex and clears the logs list
func (ml *MockLogzioListener) ClearLogsBuffer() {
	mutex.Lock()
	defer mutex.Unlock()

	// clear the logs list
	ml.LogsList = nil
}

// SetServerError is a method that locks the mutex and sets the server error to true
func (ml *MockLogzioListener) SetServerError() {
	mutex.Lock()
	defer mutex.Unlock()

	serverError = true
}

// ClearServerError is a method that locks the mutex and sets the server error to false
func (ml *MockLogzioListener) ClearServerError() {
	mutex.Lock()
	defer mutex.Unlock()

	serverError = false
}

// GetServerError is a method that locks the mutex and returns the server error status
func (pf *PersistentFlags) GetServerError() bool {
	mutex.Lock()
	defer mutex.Unlock()

	return serverError
}

// SetServerError is a method that locks the mutex and sets the server error to true
func (pf *PersistentFlags) SetServerError() {
	mutex.Lock()
	defer mutex.Unlock()

	serverError = true
}

// ClearServerError is a method that locks the mutex and sets the server error to false
func (pf *PersistentFlags) ClearServerError() {
	mutex.Lock()
	defer mutex.Unlock()

	serverError = false
}

// Thread is a struct that wraps a function to be run in a separate goroutine
type Thread struct {
	target func() // the function to be run in the goroutine
}

func (t *Thread) Start() {
	go t.target()
}

func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
func GetMockListenerURL() (mockListerURL string) {
	mockListerURL = fmt.Sprintf("http://%s:%d", MockListener.Host, MockListener.Port)
	return mockListerURL
}

// StartMockLogzioListener is a function that initializes a new mock Logzio listener
// and starts it in a separate goroutine. It also logs the URL on which the listener is running.
func StartMockLogzioListener() {
	MockListener = NewMockLogzioListener()         // create a new mock listener
	go MockListener.StartListening()               // start the listener in a goroutine
	mockListerURL := GetMockListenerURL()          // get the URL where the listener is running
	log.Printf("Listening on %s\n", mockListerURL) // log the URL
}

// SetupMockListener is a function that starts the mock Logzio listener, sets environment variables
// for the Logzio token and the listener URL, and returns a boolean indicating whether the listener
// is running successfully. It waits for the listener to be ready before setting the environment variables.
func SetupMockListener() (isListening bool) {
	// Start mock listener
	// Create a channel to signal when the listener is ready
	listenerReady := make(chan bool) // a channel to signal when the listener is ready

	// Start mock listener in a separate goroutine
	go func() {
		StartMockLogzioListener() // start the listener
		// Signal that the listener is ready
		log.Println("Mock listener started")
		listenerReady <- true // Send signal to listenerReady

	}()

	// Wait for the listener to be ready
	<-listenerReady // block until the listener is ready
	mockListener := MockListener
	if mockListener != nil {
		mockListenerURL := GetMockListenerURL()
		if mockListenerURL != "" {
			// Set the LOGZIO_TOKEN environment variable
			err := os.Setenv("LOGZIO_TOKEN", "test-shipping-token")
			if err != nil {
				log.Println("Failed to set LOGZIO_TOKEN:", err)
				return
			}
			// Set the LOGZIO_LISTENER environment variable
			err = os.Setenv("LOGZIO_LISTENER", mockListenerURL)
			if err != nil {
				log.Println("Failed to set LOGZIO_LISTENER:", err)
				return
			}
			isListening = true // indicate that the listener is running successfully
		}
	} else {
		log.Printf("Failed to start mock listener") // log a message if the listener failed to start
	}
	return isListening // return the status of the listener
}
