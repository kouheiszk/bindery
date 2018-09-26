package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"unicode"
)

const (
	Name    = "bindery"
	Version = "0.0.1"
)

func criticalError(err error) {
	logError("%s", err)
	logInfo("Run '%s --help' for usage\n", Name)
	onExit()
	os.Exit(1)
}

func onExit() {
	removeTempDirectory()
}

func main() {
	minLogLevel_ = LogLevelError

	unicode.IsLetter("0")
	fmt.Println("001" < "002")
	return

	// -----------------------------------------------------------------------------------
	// Handle SIGINT (Ctrl + C)
	// -----------------------------------------------------------------------------------

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	go func() {
		<-signalChan
		logInfo("Operation has been aborted.")
		onExit()
		os.Exit(2)
	}()
	defer onExit()

	// -----------------------------------------------------------------------------------
	// Parse arguments
	// -----------------------------------------------------------------------------------

	opts, args, err := parseArguments()
	if err != nil {
		criticalError(err)
	}
	if opts == nil {
		return
	}
	if opts.Verbose {
		minLogLevel_ = LogLevelDebug
	}

	// -----------------------------------------------------------------------------------
	// Handle version command
	// -----------------------------------------------------------------------------------

	if opts.Version {
		err = printVersion()
		if err != nil {
			criticalError(err)
		}
		return
	}

	// -----------------------------------------------------------------------------------
	// Retrieve source directories
	// -----------------------------------------------------------------------------------

	sourceDirectoryPaths, err := sourceDirectoryPathsFromArgs(args, true)
	if err != nil {
		criticalError(err)
	}
	if len(sourceDirectoryPaths) == 0 {
		criticalError(errors.New("no sources to convert pdf"))
	}

	// -----------------------------------------------------------------------------------
	// Create temporary directory
	// -----------------------------------------------------------------------------------

	tempDirectoryPath, err := prepareTempDirectory()
	if err != nil {
		criticalError(err)
	}

	// -----------------------------------------------------------------------------------
	// Convert images to pdf
	// -----------------------------------------------------------------------------------

	err = processConvertImagesToPdf(sourceDirectoryPaths, tempDirectoryPath, opts.Concurrency, opts.Dispose)
	if err != nil {
		criticalError(err)
	}
}
