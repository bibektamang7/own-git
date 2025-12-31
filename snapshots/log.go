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

type GitLog struct {
	HeadCommitHash string
	IsOneline      bool
}

func (gl *GitLog) logCommit(gitBasePath, commitHash string) error {
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

	hasParent := false
	authorIdx := 1

	if strings.HasPrefix(commitLines[1], "parent") {
		authorIdx = 2
		hasParent = true
	}
	if gl.IsOneline {
		fmt.Printf("%s", commitHash[:7])
		if commitHash == gl.HeadCommitHash {
			fmt.Printf(" (HEAD -> main) ")
		}
		fmt.Println(commitLines[len(commitLines)-1])
	} else {
		line := strings.TrimPrefix(commitLines[authorIdx], "author ")
		lt := strings.Index(line, "<")
		gt := strings.Index(line, ">")

		if lt == -1 || gt == -1 || gt < lt {
			return fmt.Errorf("malformed commit file")
		}
		name := strings.TrimSpace(line[:lt])
		email := line[lt+1 : gt]

		rest := strings.TrimSpace(line[gt+1:])
		parts := strings.Split(rest, " ")

		timestamp := parts[0]
		timezone := parts[1]

		fmt.Printf("commit %s\n", commitHash)
		if !hasParent {
			fmt.Printf("Author: %s %s\n", name, email)
			fmt.Println("Date:  ", timestamp, timezone)
			fmt.Println(commitLines[3])
			return nil
		}

		fmt.Printf("Author: %s %s\n", name, email)
		fmt.Println("Date:  ", timestamp, timezone)
		fmt.Println(commitLines[4])

	}

	if !hasParent {
		return nil
	}
	// TODO: "\n" at the end of hash
	parentParts := strings.Split(commitLines[1], " ")

	return gl.logCommit(gitBasePath, parentParts[1])

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
	onelineArg := os.Args[2]
	isOnlineArg := false
	if onelineArg == "--oneline" {
		isOnlineArg = true
	}
	gitLog := &GitLog{
		IsOneline:      isOnlineArg,
		HeadCommitHash: commitHash,
	}
	if err := gitLog.logCommit(filePath, commitHash); err != nil {
		return err
	}
	return nil
}
