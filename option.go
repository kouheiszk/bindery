package main

import (
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"strings"
)

type CommandLineOptions struct {
	Concurrency int  `short:"c" long:"concurrency" description:"Concurrency number of converting images to pdf." default:"4"`
	Dispose     bool `short:"d" long:"dispose" description:"Dispose of images and directories who was bound to pdf."`
	Verbose     bool `short:"v" long:"verbose" description:"Enable verbose output."`
	Version     bool `short:"V" long:"version" description:"Displays version information."`
}

func parseArguments() (*CommandLineOptions, []string, error) {
	var opts CommandLineOptions
	parser := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	args, err := parser.Parse()
	if err != nil {
		t := err.(*flags.Error).Type
		if t == flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
			printHelp()
			return nil, nil, nil
		} else {
			return nil, nil, err
		}
	}
	if len(args) == 0 {
		return nil, nil, errors.New("no files or directories passed")
	}

	return &opts, args, nil
}

func printHelp() {
	info := `
Examples:
  Process all the images in the specified directory:
  % PROGRAM /path/to/images/*
`
	fmt.Println(strings.Replace(info, "PROGRAM", PROGRAM, -1))
}

func printVersion() error {
	fmt.Println(VERSION)
	return nil
}
