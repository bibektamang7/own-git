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
	treePaths, err := ParseHeadAndCommitFile(fullpath)
	if err != nil {
		return err
	}

	for k, _ := range status.IndexMap {
		if _, ok := treePaths.TreePaths[k]; !ok {
			status.StagedFiles = append(status.StagedFiles, k)
		}
	}

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
