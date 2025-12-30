package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bibektamang7/own-git/snapshots"
)

const (
	INIT     string = "init"
	STATUS   string = "status"
	COMMIT   string = "commit"
	ADD      string = "add"
	LOG      string = "log"
	CAT_FILE string = "cat-file"
)

func main() {
	// commands := os.Args[1:]
	if len(os.Args[1:]) < 1 {
		log.Fatal("commands required")
	}
	if os.Args[1] == "" {
		log.Fatal("empty command")
	}

	switch os.Args[1] {
	case INIT:
		if err := snapshots.InitializeGit(); err != nil {
			log.Fatal("INIT COMMAND ERROR: ", err)
		}
		fmt.Println("Initialized Git Successfully")
	case STATUS:
		if err := snapshots.HandleStatusCommand(); err != nil {
			log.Fatal("STATUS COMMAND ERROR: ", err)
		}
		fmt.Println("Git status command")
	case COMMIT:
		if err := snapshots.HandleCommitCommand(); err != nil {
			log.Fatal(err)
		}
	case ADD:
		if err := snapshots.HandleAddCommand(); err != nil {
			log.Fatal("ADD COMMAND ERROR: ", err)
		}
		fmt.Println("Git add command")
	case LOG:
		fmt.Println("git log command")
	case CAT_FILE:
		if err := snapshots.HandleCatFile(); err != nil {
			log.Fatal("CAT FILE ERROR: ", err)
		}
		fmt.Println("git cat file command")
	default:
		log.Fatal("invalid command arguments")
	}

}
