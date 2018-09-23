package main

import (
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
)

const (
	PROGRAM = "bindery"
	VERSION = "0.0.1"
)

func criticalError(err error) {
	logError("%s", err)
	logInfo("Run '%s --help' for usage\n", PROGRAM)
	onExit()
	os.Exit(1)
}

func onExit() {
	removeTempDirectory()
}

func targetDirectoryPathsFromArgs(inputs []string, recursive bool) ([]string, error) {
	var paths []string

	// Convert inputs to absolute paths
	for _, arg := range inputs {
		if strings.Index(arg, "*") < 0 {
			absolutePath, err := filepath.Abs(arg)
			if err != nil {
				logDebug("Invalid argument", arg)
				continue
			}
			paths = append(paths, absolutePath)
			continue
		}
		matches, err := filepath.Glob(arg)
		if err != nil {
			return []string{}, err
		}
		children, err := targetDirectoryPathsFromArgs(matches, false)
		if err != nil {
			return []string{}, err
		}
		paths = append(paths, children...)
	}

	// Get one directory up path if all paths are images
	if isOnlySupportedImages(paths) {
		directory, _ := filepath.Split(paths[0])
		return targetDirectoryPathsFromArgs([]string{directory}, false)
	}

	// Get directory from paths and check exists images in the directories
	var targetDirectoryPaths []string
	for _, path := range paths {
		state, err := os.Stat(path)
		if err != nil || !state.IsDir() {
			continue
		}
		children, err := filepath.Glob(filepath.Join(path, "*"))
		if err != nil {
			continue
		}
		if isOnlySupportedImages(children) {
			targetDirectoryPaths = append(targetDirectoryPaths, path)
		} else if recursive {
			children, err = targetDirectoryPathsFromArgs([]string{filepath.Join(path, "*")}, false)
			if err != nil {
				continue
			}
			targetDirectoryPaths = append(targetDirectoryPaths, children...)
		}
	}

	sort.Strings(targetDirectoryPaths)

	return targetDirectoryPaths, nil
}

func main() {
	minLogLevel_ = 0

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
		minLogLevel_ = 0
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
	// Retrieve target directories
	// -----------------------------------------------------------------------------------

	targetDirectoryPaths, err := targetDirectoryPathsFromArgs(args, true)
	if err != nil {
		criticalError(err)
	}
	if len(targetDirectoryPaths) == 0 {
		criticalError(errors.New("no targets to convert pdf"))
	}

	// -----------------------------------------------------------------------------------
	// Create temporary directory
	// -----------------------------------------------------------------------------------

	tempDirectory, err := prepareTempDirectory()
	if err != nil {
		criticalError(err)
	}

	// -----------------------------------------------------------------------------------
	// Convert images to pdf
	// -----------------------------------------------------------------------------------

	err = processConvertImagesToPdf(targetDirectoryPaths, tempDirectory, opts.Concurrency)
	if err != nil {
		criticalError(err)
	}

	// -----------------------------------------------------------------------------------
	// Move converted pdfs and dispose target directories if needed
	// -----------------------------------------------------------------------------------

}
