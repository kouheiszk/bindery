package main

import (
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/b4b4r07/gomi/darwin"
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

func sourceDirectoryPathsFromArgs(inputs []string, recursive bool) ([]string, error) {
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
		children, err := sourceDirectoryPathsFromArgs(matches, false)
		if err != nil {
			return []string{}, err
		}
		paths = append(paths, children...)
	}

	// Get one directory up path if all paths are images
	if isOnlySupportedImages(paths) {
		directory, _ := filepath.Split(paths[0])
		return sourceDirectoryPathsFromArgs([]string{directory}, false)
	}

	// Get directory from paths and check exists images in the directories
	var sourceDirectoryPaths []string
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
			sourceDirectoryPaths = append(sourceDirectoryPaths, path)
		} else if recursive {
			children, err = sourceDirectoryPathsFromArgs([]string{filepath.Join(path, "*")}, false)
			if err != nil {
				continue
			}
			sourceDirectoryPaths = append(sourceDirectoryPaths, children...)
		}
	}

	sort.Strings(sourceDirectoryPaths)

	return sourceDirectoryPaths, nil
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

	tempDirectory, err := prepareTempDirectory()
	if err != nil {
		criticalError(err)
	}

	// -----------------------------------------------------------------------------------
	// Convert images to pdf
	// -----------------------------------------------------------------------------------

	pdfs, err := processConvertImagesToPdf(sourceDirectoryPaths, tempDirectory, opts.Concurrency)
	if err != nil {
		criticalError(err)
	}

	// -----------------------------------------------------------------------------------
	// Move converted pdfs
	// -----------------------------------------------------------------------------------

	for _, pdf := range pdfs {
		destPath := filepath.Join(filepath.Dir(pdf.SourceDirectoryPath), filepath.Base(pdf.Path))

		// Filename duplicate guard
		extension := filepath.Ext(destPath)
		destPathBase := destPath[0:strings.Index(destPath, extension)]
		i := 0
		for {
			_, err := os.Stat(destPath)
			if err != nil {
				break
			}
			i += 1
			destPath = destPathBase + " (" + strconv.Itoa(i) + ")" + extension
		}

		// Move pdf to source directory's directory from temporary directory
		if err := os.Rename(pdf.Path, destPath); err != nil {
			criticalError(err)
		}
	}

	// -----------------------------------------------------------------------------------
	// Dispose source directories if needed
	// -----------------------------------------------------------------------------------

	if opts.Dispose {
		for _, pdf := range pdfs {
			_, err := darwin.Trash(pdf.SourceDirectoryPath)
			if err != nil {
				if err := os.RemoveAll(pdf.SourceDirectoryPath); err != nil {
					logError("%s", err)
				}
			}
		}
	}
}
