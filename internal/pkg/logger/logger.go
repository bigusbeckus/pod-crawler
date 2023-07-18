package logger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/config"
)

var Info *log.Logger
var Warn *log.Logger
var Error *log.Logger

// Creates a file writer with predetermined options given a filename
func createFileWriter(filename string) (io.Writer, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Creates a file transport that still outputs to stdout
func getLogfileName(logDestination string, tag string) string {
	// Generate log file name
	logFile := logDestination

	timeStr := time.Now().Format(time.RFC3339)
	timeStr = strings.Split(timeStr, "T")[0]
	logFile = logFile + timeStr

	if tag != "" {
		logFile = fmt.Sprintf("%s-%s", logFile, tag)
	}
	logFile = logFile + ".log"

	return logFile
}

func verifyDestination(logsDirectory string) error {
	_, err := os.Stat(logsDirectory)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(logsDirectory, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func Init() error {
	logDestination := config.AppConfig.LogDestination

	err := verifyDestination(logDestination)
	if err != nil {
		return err
	}

	// Get filenames
	defaultFilename := getLogfileName(logDestination, "")
	errorFilename := getLogfileName(logDestination, "errors")

	// Create files
	defaultFile, defaultFileErr := createFileWriter(defaultFilename)
	errorFile, errorFileErr := createFileWriter(errorFilename)
	if err := errors.Join(defaultFileErr, errorFileErr); err != nil {
		return err
	}

	// Create loggers
	colorStdout := &colorWriter{
		out: os.Stdout,
	}

	Info = log.New(
		io.MultiWriter(defaultFile, colorStdout),
		infoPrefix,
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	Warn = log.New(
		io.MultiWriter(defaultFile, colorStdout),
		warnPrefix,
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	Error = log.New(
		io.MultiWriter(defaultFile, errorFile, colorStdout),
		errorPrefix,
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	return nil
}
