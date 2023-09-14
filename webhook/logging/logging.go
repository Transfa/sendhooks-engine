package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// logging helps with logging in the application. Logs are saved in a daily YYYY-MM-DD.log file format.
// Two types of log messages are supported: ERROR and WARNING.
// Each log is formatted as: ERROR_TYPE - YYYY-MM-DD HH:MM:SS - ERROR_STRING

const (
	ErrorType   = "ERROR"
	WarningType = "WARNING"
)

// currentDate retrieves the current date in "YYYY-MM-DD" format.
func currentDate() string {
	return time.Now().Format("2006-01-02")
}

// currentDateTime retrieves the current date and time in "YYYY-MM-DD HH:MM:SS" format.
func currentDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

var WebhookLogger = func(errorType string, errorMessage error) error {
	logFileDate := currentDate()
	logFileName := fmt.Sprintf("%s.log", logFileDate)

	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Print("Failed to open log file: \n", err)
		return err
	}
	defer file.Close()

	multi := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multi)

	logEntry := fmt.Sprintf("%s - %s - %s\n", errorType, currentDateTime(), errorMessage)
	_, err = log.Writer().Write([]byte(logEntry))
	return err
}
