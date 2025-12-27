package snapshots

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Status struct {
	IndexMap       map[string]IndexLine
	seen           map[string]bool
	baseRoot       string
	UntractedFiles []string
	ModifiedFiles  []string
	StagedFiles    []string // git compares previous commit i.e HEAD
	DeletedFiles   []string
}

func NewStatus() *Status {
	return &Status{
		UntractedFiles: []string{},
		ModifiedFiles:  []string{},
		StagedFiles:    []string{},
		DeletedFiles:   []string{},
		IndexMap:       make(map[string]IndexLine),
		seen:           make(map[string]bool),
		baseRoot:       "",
	}
}

func (s *Status) parseIndexFile() error {
	path := s.baseRoot + ROOTDIR + "index"
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) != 5 {
			return fmt.Errorf("corrupt index line: %q", line)
		}

		mode, err := strconv.ParseUint(parts[2], 8, 32)
		if err != nil {
			return err
		}
		size, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			return err
		}
		ts, err := strconv.ParseInt(parts[4], 10, 64)
		if err != nil {
			return err
		}

		idx := IndexLine{
			Fullpath:   parts[0],
			BlobHash:   parts[1],
			FileMode:   uint32(mode),
			FileSize:   size,
			TimeStamps: ts,
		}

		s.IndexMap[idx.Fullpath] = idx
	}

	return nil
}

func (s *Status) visitWorkingDirFiles(repoRoot string) error {
	return filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".owngit" {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(s.baseRoot, path)
		if err != nil {
			return err
		}

		s.seen[rel] = true

		if idxLine, ok := s.IndexMap[rel]; ok {

			if idxLine.FileSize != info.Size() ||
				idxLine.TimeStamps != info.ModTime().UnixNano() {
				s.ModifiedFiles = append(s.ModifiedFiles, idxLine.Fullpath)
			}
			return nil
		}

		// New file
		s.UntractedFiles = append(s.UntractedFiles, rel)
		return nil
	})
}

func (s *Status) deletedFiles() {
	for k, _ := range s.IndexMap {
		if _, ok := s.seen[k]; !ok {
			s.DeletedFiles = append(s.DeletedFiles, k)
		}
	}
}

var ERROR_MALFORMED_TREE_FORMAT = fmt.Errorf("malformed tree format")

func (s *Status) parseCommitFile(r io.Reader, basePath string) error {
	treeReader := bufio.NewReader(r)

	treeLine, err := treeReader.ReadString('\n')
	if err != nil {
		return err
	}

	treeParts := strings.SplitN(treeLine, " ", 2)
	if len(treeParts) != 2 {
		return ERROR_MALFORMED_COMMIT_FORMAT
	}
	if treeParts[0] != "tree" {
		return ERROR_MALFORMED_COMMIT_FORMAT
	}
	treeHash := treeParts[1]
	treeHashParts := strings.SplitN(treeHash, treeHash[:2], 2)
	if len(treeHashParts) != 2 {
		return ERROR_MALFORMED_COMMIT_FORMAT
	}
	treeHashFilePath := basePath + ROOTDIR + "objects/" + treeHashParts[0] + "/" + treeHashParts[1]
	treeHashFile, err := os.Open(treeHashFilePath)
	if err != nil {
		return err
	}
	defer treeHashFile.Close()

	treePaths := NewTreePaths()
	if err := treePaths.parseTreeFile(treeHashFile, basePath, ""); err != nil {
		return err
	}

	for k, _ := range s.IndexMap {
		if _, ok := treePaths.TreePaths[k]; !ok {
			s.StagedFiles = append(s.StagedFiles, k)
		}
	}
	return nil
}

var ERROR_MALFORMED_COMMIT_FORMAT = fmt.Errorf("malformed commit format")

func (s *Status) parseHeadFile(basePath string) error {
	fi, err := os.Open(basePath + ROOTDIR + "HEAD")
	if err != nil {
		return err
	}
	defer fi.Close()

	reader := bufio.NewReader(fi)

	commitHash, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	parts := strings.SplitN(commitHash, commitHash[:2], 2)
	if len(parts) != 2 {
		return ERROR_MALFORMED_COMMIT_FORMAT
	}
	commitFilePath := basePath + ROOTDIR + "objects/" + parts[0] + "/" + parts[1]
	commitFile, err := os.Open(commitFilePath)
	if err != nil {
		return err
	}
	defer commitFile.Close()
	s.parseCommitFile(commitFile, basePath)

	return nil
}

func HandleStatusCommand() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	fullpath, ok, err := CheckGitFolderExists(path)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("couldn't found .owngit folder")
	}

	status := NewStatus()
	status.baseRoot = fullpath
	if err := status.parseIndexFile(); err != nil {
		return err
	}
	if err := status.visitWorkingDirFiles(fullpath); err != nil {
		return err
	}
	status.deletedFiles()

	fmt.Println("On branch main")

	if len(status.ModifiedFiles) < 1 && len(status.StagedFiles) < 1 &&
		len(status.UntractedFiles) < 1 && len(status.DeletedFiles) < 1 {
		fmt.Println("nothing to commit, working tree clean")
		return nil
	}
	if len(status.UntractedFiles) > 0 {
		fmt.Println("Untracked files:")
		fmt.Println("\t(use \"git add <file>...\" to include in what will be commited):")
		for _, path := range status.UntractedFiles {
			fmt.Printf("\t\t%s\n", path)
		}
	}

	if len(status.ModifiedFiles) > 0 {
		fmt.Println("Modified files:")
		for _, path := range status.ModifiedFiles {
			fmt.Printf("\t\t%s\n", path)
		}
	}
	if len(status.DeletedFiles) > 0 {
		fmt.Println("Deleted files:")
		for _, path := range status.DeletedFiles {
			fmt.Printf("\t\t%s\n", path)
		}
	}
	if len(status.StagedFiles) > 0 {
		fmt.Println("changes to be committed:")
		for _, path := range status.StagedFiles {
			fmt.Printf("\t\t%s\n", path)
		}
	} else if len(status.UntractedFiles) > 0 {
		fmt.Println("nothing added to commit but untracked files present (use \"git add\" to track)")
	}

	return nil
}
