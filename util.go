package main

import (
	"github.com/b4b4r07/gomi/darwin"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

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

func destPathFromSourceDirectoryPath(directoryPath string) string {
	// Declare destination path
	destPath := filepath.Join(filepath.Dir(directoryPath), filepath.Base(directoryPath)+".pdf")

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

	return destPath
}

func trashDirectory(directoryPath string) error {
	_, err := darwin.Trash(directoryPath)
	if err != nil {
		if err := os.RemoveAll(directoryPath); err != nil {
			return err
		}
	}

	return nil
}
