package main

import (
	"fmt"
	"log"
	"os"
)

const (
	INIT   string = "init"
	STATUS string = "status"
	COMMIT string = "commit"
	LOG    string = "log"
)

func main() {
	commands := os.Args[1:]
	if len(commands) < 1 {
		log.Fatal("commands required")
	}
	if commands[0] == "" {
		log.Fatal("empty command")
	}

	switch commands[0] {
	case INIT:
		fmt.Println("git init command")
	case STATUS:
		fmt.Println("git status command")
	case COMMIT:
		fmt.Println("git commit command")
	case LOG:
		fmt.Println("git log command")
	default:
		log.Fatal("invalid command arguments")
	}

}
