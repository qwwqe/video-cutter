package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsageAndQuit()
	}

	var command func([]string) error
	switch os.Args[1] {
	case "edl":
		command = GenerateRandomEDL
	case "cover":
		command = GenerateRandomCoverImage
	}

	if command == nil {
		fmt.Printf("Unknown command \"%s\"\n", os.Args[1])
		printUsageAndQuit()
	}

	if err := command(os.Args[2:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func printUsageAndQuit() {
	fmt.Println("Usage: vc <command> [flags]")
	fmt.Println("Valid commands: edl cover")
	os.Exit(1)
}
