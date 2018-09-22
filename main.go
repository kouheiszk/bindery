package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jessevdk/go-flags"
)

var flagParser_ *flags.Parser

const (
	APPNAME = "bindery"
)

type CommandLineOptions struct {
	Dispose bool `short:"d" long:"dispose" description:"Dispose of images and directories who was bound to pdf."`
	Verbose bool `short:"v" long:"verbose" description:"Enable verbose output."`
	Version bool `short:"V" long:"version" description:"Displays version information."`
}

func printHelp() {
	flagParser_.WriteHelp(os.Stdout)

	info := `
Examples:
  Process all the images in the specified directory:
  % APPNAME /path/to/images/*
`
	fmt.Println(strings.Replace(info, "APPNAME", APPNAME, -1))
}

func criticalError(err error) {
	logError("%s", err)
	logInfo("Run '%s --help' for usage\n", APPNAME)
	os.Exit(1)
}

func onExit() {
}

func targetDirectoryPathsFromArgs(args []string, recursive bool) ([]string, error) {
	var paths []string

	// Convert args to absolute paths
	for _, arg := range args {
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
		}
		if recursive {
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

	var opts CommandLineOptions
	flagParser_ = flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	args, err := flagParser_.Parse()
	if err != nil {
		t := err.(*flags.Error).Type
		if t == flags.ErrHelp {
			printHelp()
			return
		} else {
			criticalError(err)
		}
	}

	if opts.Verbose {
		minLogLevel_ = 0
	}

	if len(args) == 0 {
		criticalError(errors.New("no files or directories passed"))
	}

	// -----------------------------------------------------------------------------------
	// Handle selected command
	// -----------------------------------------------------------------------------------

	var commandName string
	if opts.Version {
		commandName = "version"
	} else {
		commandName = "bind"
	}

	var commandErr error
	switch commandName {
	case "version":
		commandErr = handleVersionCommand(&opts, args)
	}

	if commandErr != nil {
		criticalError(commandErr)
	}

	if commandName != "bind" {
		return
	}

	targetDirectories, err := targetDirectoryPathsFromArgs(args, true)

	if err != nil {
		criticalError(err)
	}

	if len(targetDirectories) == 0 {
		criticalError(errors.New("no targets to convert pdf"))
	}

	fmt.Println(targetDirectories)
}
