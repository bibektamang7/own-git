package snapshots

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type ContentType string

const (
	Blob ContentType = "blob"
	Tree ContentType = "tree"
)

type CommitTree struct {
	fileMode    os.FileMode
	contentType ContentType
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

type TreePaths struct {
	TreePaths map[string]string
}

func NewTreePaths() *TreePaths {
	return &TreePaths{
		TreePaths: make(map[string]string),
	}
}

func (t *TreePaths) parseTreeFile(r io.Reader, gitRoot, treePath string) error {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) != 4 {
			return ERROR_MALFORMED_TREE_FORMAT
		}
		filePath := treePath + parts[3]
		if parts[1] == "tree" {
			treeHash := parts[1]
			treeHashParts := strings.SplitN(treeHash, treeHash[:2], 2)
			if len(treeHashParts) != 2 {
				return ERROR_MALFORMED_COMMIT_FORMAT
			}
			treeHashFilePath := gitRoot + ROOTDIR + "objects/" + treeHashParts[0] + "/" + treeHashParts[1]
			treeHashFile, err := os.Open(treeHashFilePath)
			if err != nil {
				return err
			}
			if err := t.parseTreeFile(treeHashFile, gitRoot, filePath); err != nil {
				return err
			}
			defer treeHashFile.Close()
			continue
		}

		// assigning blob to tree path, which is useful
		// when comparing with index lines
		t.TreePaths[filePath] = parts[2]
	}
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
