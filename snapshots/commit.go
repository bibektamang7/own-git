package snapshots

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type ContentType string

const (
	Blob ContentType = "blob"
	Tree ContentType = "tree"
)

type CommitTree struct {
	contentType ContentType
	fileMode    os.FileMode
	Hash        string
	Name        string
}

type Commit struct {
	tree     string
	parent   string
	author   []string
	commiter []string
	message  string
}

func compareAndFindStagedFiles() error {
	return nil
}

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

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Println("message: ", *msg)

	_, ok, err := CheckGitFolderExists(path)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("outside of Git repository")
	}

	return nil
}
