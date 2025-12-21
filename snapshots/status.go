package snapshots

import (
	"fmt"
	"os"
)

type Status struct {
	UntractedFiles []string
	ModifiedFiles  []string
	StagedFiles    []string
	DeletedFiles   []string
}

func NewStatus() *Status {
	return &Status{
		UntractedFiles: []string{},
		ModifiedFiles:  []string{},
		StagedFiles:    []string{},
		DeletedFiles:   []string{},
	}
}

/*
	steps:
		1: checking index file, if files path not found in index -> untracked
		2: if file in the index, not found in working tree -> deleted
		3: comparing modified data and size of each files,
		4: if found difference, then hash content and compare from index's hash
		5: if diff -> modified
		6: then StagedFiles | trackedFiles
*/

func (s *Status) checkIndexFile() error {
	return nil
}
func HandleStatusCommand() error {

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	_, ok, err := CheckGitFolderExists(path)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("couldn't found .owngit folder")
	}

	status := NewStatus()
	status.checkIndexFile()

	if err := status.checkIndexFile(); err != nil {
		return err
	}
	return nil
}
