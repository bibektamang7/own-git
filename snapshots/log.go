package snapshots

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var ERROR_OUTSIDE_GIT = fmt.Errorf("outside git repository")

func logCommit(gitBasePath, commitHash string) error {
	commitFilePath := filepath.Join(gitBasePath, ROOTDIR, "objects", commitHash[:2], commitHash[2:])
	fi, err := os.Open(commitFilePath)
	if err != nil {
		return err
	}
	defer fi.Close()

	var commitLines []string

	reader := bufio.NewReader(fi)

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if line == "\n" {
			continue
		}
		commitLines = append(commitLines, line)
	}
	fmt.Printf("commit %s\n", commitHash)
	hasParent := false

	parentParts := strings.Split(commitLines[1], " ")

	if parentParts[0] == "parent" {
		hasParent = true
	}

	if !hasParent {
		fmt.Printf("Author: %s %s", parentParts[1], parentParts[2])
		fmt.Println("Date:  ", parentParts[3], parentParts[4])
		fmt.Println(commitLines[3])
		return nil
	}

	authorParts := strings.Split(commitLines[2], " ")

	fmt.Printf("Author: %s %s", authorParts[1], authorParts[2])
	fmt.Println("Date:  ", parentParts[3], parentParts[4])
	fmt.Println(commitLines[4])

	return logCommit(gitBasePath, parentParts[1])

}

func HandleLog() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	filePath, ok, err := CheckGitFolderExists(path)
	if err != nil {
		return err
	}
	if !ok {
		return ERROR_OUTSIDE_GIT
	}

	commitHash, err := GetPreviousCommitHash(filePath)
	if err != nil {
		return err
	}
	if err := logCommit(filePath, commitHash); err != nil {
		return err
	}
	return nil
}
