package main

import (
	"fmt"
	"os"
	"strings"
)

const ROOTDIR string = "/.owngit"

var FILES = []string{
	"HEAD",      // Current branch or commit
	"ORIG_HEAD", // Backup of previous state for undoing operations
}
var FOLDERS = []string{
	"hooks", // Scripts triggered by Git events
	"info",  // Local repo metadata (e.g. excludes)
	"logs",  // History of reference changes (reflog)
}

var ERROR_CHECK_FOLDER_EXISTS = fmt.Errorf("failed on checking existing folder")

func checkGitFolderExists(path string) (string, bool, error) {
	if path == "" {
		return "", false, ERROR_CHECK_FOLDER_EXISTS
	}
	parts := strings.Split(path, "/")
	numParts := len(parts)
	if numParts < 1 {
		return "", false, ERROR_CHECK_FOLDER_EXISTS
	}
	for i := numParts; i >= 0; i-- {
		// could improve : That's for later
		currentPath := strings.Join(parts[:i], "/")
		folder := currentPath + ROOTDIR
		f, err := os.Stat(folder)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", false, err
		}
		if f.IsDir() {
			return currentPath, true, nil
		}
	}

	return "", false, nil
}

func InitializeGit() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	existPath, ok, err := checkGitFolderExists(path)
	if err != nil {
		return err
	}
	if ok {
		// TODO:
		// handle reinitializing git
		fmt.Println("Reinitializing Git to ", existPath)
		return nil
	}
	fmt.Println("Initializing Git... ")
	// create root .owngit folder
	fullPath := path + ROOTDIR
	err = os.Mkdir(fullPath, os.ModePerm)
	return err
}
