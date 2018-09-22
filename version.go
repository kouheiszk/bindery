package main

import "fmt"

const VERSION = "0.0.1"

func handleVersionCommand(opts *CommandLineOptions, args []string) error {
	fmt.Println(VERSION)
	return nil
}
