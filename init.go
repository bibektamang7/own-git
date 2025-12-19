package main

import (
	"fmt"
	"os"
	"strings"
)

const ROOTDIR string = "/.owngit/"

var DEFAULT_CONFIGS = map[string]string{
	"filemode":        "false",
	"bare":            "false",
	"localrefupdates": "true",
}

var FILES = []string{
	"HEAD",      // Current branch or commit
	"ORIG_HEAD", // Backup of previous state for undoing operations
	"index",     // fast lookup
	"config",    // local git config
}
var FOLDERS = []string{
	"hooks",   // Scripts triggered by Git events
	"info",    // Local repo metadata (e.g. excludes)
	"logs",    // History of reference changes (reflog)
	"objects", // commits hash
	"refs/heads",
	// "refs/remotes", // add only when remote is set
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

func InitializeFoldersAndFiles(path string) error {
	// create root .owngit folder

	for _, folder := range FOLDERS {
		if err := os.MkdirAll(path+folder, os.ModePerm); err != nil {
			return err
		}
	}
	for _, file := range FILES {
		// TODO: if file == "config" -> initialize default config
		// i.e INI file with default key-value
		if _, err := os.Create(path + file); err != nil {
			return err
		}
	}
	return nil
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

	fullPath := path + ROOTDIR
	if err := InitializeFoldersAndFiles(fullPath); err != nil {
		return err
	}

	fmt.Println("Initialized empty Git repository in ", fullPath)
	return nil
}
