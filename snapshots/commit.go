package snapshots

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func HandleCommitCommand() error {
	fs := flag.NewFlagSet("commit", flag.ExitOnError)
	msg := fs.String("m", "", "commit message")

	fs.Parse(os.Args[2:])
	args := fs.Args()

	if len(args) > 0 {
		log.Fatalf("invalid command argument: %s\n", args[0])
	}
	if len(*msg) < 1 {
		return fmt.Errorf("empty commit message")
	}

	fmt.Println("message: ", *msg)

	return nil
}
