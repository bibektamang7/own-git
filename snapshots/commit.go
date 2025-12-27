package snapshots

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

var (
	ERROR_MALFORMED_TREE_FORMAT   = fmt.Errorf("malformed tree format")
	ERROR_MALFORMED_COMMIT_FORMAT = fmt.Errorf("malformed commit format")
)

func NewTreePaths() TreePaths {
	return TreePaths{
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
			treeHash := strings.TrimSpace(parts[1])
			treeHashParts := []string{treeHash[:2], treeHash[2:]}
			if len(treeHashParts) != 2 {
				return ERROR_MALFORMED_COMMIT_FORMAT
			}
			// treeHashFilePath := gitRoot + ROOTDIR + "objects/" + treeHashParts[0] + "/" + treeHashParts[1]
			// TODO: IF IT WORKS THEN CHANGE OTHER TOO
			treeHashFilePath := filepath.Join(gitRoot, ROOTDIR, "objects", treeHashParts[0], treeHashParts[1])
			treeHashFile, err := os.Open(treeHashFilePath)
			if err != nil {
				return err
			}

			defer treeHashFile.Close()
			if err := t.parseTreeFile(treeHashFile, gitRoot, filePath); err != nil {
				return err
			}
			continue
		}

		// assigning blob to tree path, which is useful
		// when comparing with index lines
		t.TreePaths[filePath] = parts[2]
	}
}

func parseCommitFile(r io.Reader, basePath string) (TreePaths, error) {
	treeReader := bufio.NewReader(r)

	treeLine, err := treeReader.ReadString('\n')
	if err != nil {
		return TreePaths{}, err
	}

	treeParts := strings.SplitN(treeLine, " ", 2)
	if len(treeParts) != 2 {
		return TreePaths{}, ERROR_MALFORMED_COMMIT_FORMAT
	}
	if treeParts[0] != "tree" {
		return TreePaths{}, ERROR_MALFORMED_COMMIT_FORMAT
	}
	treeHash := strings.TrimSpace(treeParts[1])
	treeHashParts := []string{treeHash[:2], treeHash[2:]}
	if len(treeHashParts) != 2 {
		return TreePaths{}, ERROR_MALFORMED_COMMIT_FORMAT
	}
	treeHashFilePath := basePath + ROOTDIR + "objects/" + treeHashParts[0] + "/" + treeHashParts[1]
	treeHashFile, err := os.Open(treeHashFilePath)
	if err != nil {
		return TreePaths{}, err
	}
	defer treeHashFile.Close()

	treePaths := NewTreePaths()
	if err := treePaths.parseTreeFile(treeHashFile, basePath, ""); err != nil {
		return TreePaths{}, err
	}
	return treePaths, nil
}

func ParseHeadFile(basePath string) (TreePaths, error) {
	fi, err := os.Open(basePath + ROOTDIR + "HEAD")
	if err != nil {
		return TreePaths{}, err
	}
	defer fi.Close()

	reader := bufio.NewReader(fi)

	commitHash, err := reader.ReadString('\n')
	if err != nil {
		return TreePaths{}, err
	}

	commitTrimHash := strings.TrimSpace(commitHash)
	parts := []string{commitTrimHash[:2], commitTrimHash[2:]}
	if len(parts) != 2 {
		return TreePaths{}, ERROR_MALFORMED_COMMIT_FORMAT
	}
	commitFilePath := basePath + ROOTDIR + "objects/" + parts[0] + "/" + parts[1]
	commitFile, err := os.Open(commitFilePath)
	if err != nil {
		return TreePaths{}, err
	}
	defer commitFile.Close()
	return parseCommitFile(commitFile, basePath)
}

func compareAndFindStagedFiles(gitRootPath string) error {
	status := NewStatus()
	status.baseRoot = gitRootPath
	if err := status.parseIndexFile(); err != nil {
		return err
	}

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
