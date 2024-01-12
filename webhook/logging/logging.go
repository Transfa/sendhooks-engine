package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	ErrorType   = "ERROR"
	WarningType = "WARNING"
	EventType   = "EVENT"
)

// Create a new Logrus Logger
var logger = logrus.New()

// currentDate retrieves the current date in "YYYY-MM-DD" format.
func currentDate() string {
	return time.Now().Format("2006-01-02")
}

// currentDateTime retrieves the current date and time in "YYYY-MM-DD HH:MM:SS" format.
func currentDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

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

	logFileDate := currentDate()
	logFileName := fmt.Sprintf("%s.log", logFileDate)

	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Print("Failed to open log file: \n", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Print("Failed to close log file: \n", err)
		}
	}(file)

	logger.SetOutput(file)
	logger.SetFormatter(&logrus.TextFormatter{})

	// Log the entry
	switch errorType {
	case ErrorType:
		logger.WithFields(logrus.Fields{
			"date": currentDateTime(),
		}).Error(messageString)
	case WarningType:
		logger.WithFields(logrus.Fields{
			"date": currentDateTime(),
		}).Warning(messageString)
	case EventType:
		logger.WithFields(logrus.Fields{
			"date": currentDateTime(),
		}).Info(messageString)
	}

	return nil
}
