package main

import (
	"io/ioutil"
	"os"
)

var tempDirectory_ string

func prepareTempDirectory() (string, error) {
	dir, err := ioutil.TempDir("", Name)
	if err != nil {
		return dir, err
	}
	logDebug("Create temporary directory: %s", dir)
	tempDirectory_ = dir
	return dir, nil
}

func removeTempDirectory() error {
	_, err := os.Stat(tempDirectory_)
	if err != nil {
		return nil
	}

	err = os.RemoveAll(tempDirectory_)
	if err != nil {
		return err
	}

	logDebug("Removed temporary directory: %s", tempDirectory_)
	return nil
}
