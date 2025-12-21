package snapshots 

import (
	"fmt"
	"os"
	"strings"

	"github.com/bibektamang7/own-git/ini"
)

const ROOTDIR string = ".owngit/"

var DEFAULTCONFIGS = []string{
	"core.filemode=false",
	"core.bare=false",
	"core.localrefupdates=true",
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

func CheckGitFolderExists(path string) (string, bool, error) {
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
		fi, err := os.Create(path + file)
		if err != nil {
			return err
		}
		defer fi.Close()
		if file == "config" {
			fINI := ini.NewFileINI()
			for _, config := range DEFAULTCONFIGS {
				parts := strings.SplitN(config, "=", 2)
				if len(parts) != 2 {
					continue
				}
				segments := strings.SplitN(parts[0], ".", 2)
				if len(segments) != 2 {
					continue
				}
				fINI.Add(segments[0], segments[1], parts[1])
			}
			return fINI.Write(fi)
		}

	}
	return nil
}

func InitializeGit() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	existPath, ok, err := CheckGitFolderExists(path)
	if err != nil {
		return err
	}
	if ok {
		// TODO:
		// handle reinitializing git
		fmt.Println("Reinitializing Git to ", existPath)
		return nil
	}

	fmt.Println("path :", path)
	fullPath := path + ROOTDIR
	if err := InitializeFoldersAndFiles(fullPath); err != nil {
		return err
	}

	fmt.Println("Initialized empty Git repository in ", fullPath)
	return nil
}
