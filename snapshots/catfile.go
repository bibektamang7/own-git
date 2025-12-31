package snapshots

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func HandleCatFile() error {
	hash := os.Args[2]
	if len(hash) < 5 {
		return fmt.Errorf("at least 5 hash characters")
	}
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	gitRootPath, ok, err := CheckGitFolderExists(path)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("outside git repository")
	}

	hashPath := filepath.Join(gitRootPath, ROOTDIR, "objects", hash[:2], hash[2:])
	matches, err := filepath.Glob(fmt.Sprintf("%s*", hashPath))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return fmt.Errorf("no file found matching pattern : %s", hashPath)
	}
	fi, err := os.Open(matches[0])
	if err != nil {
		return err
	}
	defer fi.Close()

	reader := bufio.NewReader(fi)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("%s", line)

	}
	return nil
}
