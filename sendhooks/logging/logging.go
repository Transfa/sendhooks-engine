package logging

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	ErrorType   = "ERROR"
	WarningType = "WARNING"
	EventType   = "EVENT"
)

// Logger setup
var (
	logger      = logrus.New()
	logFile     *os.File
	logFileName string
	logMutex    sync.Mutex
)

func init() {
	setupLogFile()
	go rotateLogFileDaily()
}

// setupLogFile initializes the log file.
func setupLogFile() {
	var err error
	logFileName = currentDate() + ".log"
	logFile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	logger.SetOutput(logFile)
	logger.SetFormatter(&logrus.TextFormatter{})
}

// currentDate retrieves the current date in "YYYY-MM-DD" format.
func currentDate() string {
	return time.Now().Format("2006-01-02")
}

// currentDateTime retrieves the current date and time in "YYYY-MM-DD HH:MM:SS" format.
func currentDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// rotateLogFileDaily handles daily log rotation.
func rotateLogFileDaily() {
	for {
		time.Sleep(24 * time.Hour)
		logMutex.Lock()
		if currentDate() != logFileName[:10] {
			logFile.Close()
			setupLogFile()
		}
		logMutex.Unlock()
	}
}

// WebhookLogger logs messages with different types (Error, Warning, Event).
var WebhookLogger = func(errorType string, message interface{}) error {
	var messageString string

	switch v := message.(type) {
	case error:
		messageString = v.Error()
	case string:
		messageString = v
	default:
		return fmt.Errorf("unsupported message type: %T", message)
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	// Log the entry
	entry := logger.WithFields(logrus.Fields{
		"date": currentDateTime(),
	})
	switch errorType {
	case ErrorType:
		entry.Error(messageString)
	case WarningType:
		entry.Warning(messageString)
	case EventType:
		entry.Info(messageString)
	}

	return nil
}
