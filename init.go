package main

import (
	"fmt"
	"os"
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

func checkExists() (bool, error){
	return false, nil
}

func InitializeGit() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	// create root .owngit folder
	fullPath := path + ROOTDIR
	err = os.Mkdir(fullPath, os.ModePerm)
	fmt.Println("check: ", path)
	fmt.Println("foldre created" , fullPath)
	return err 
}
