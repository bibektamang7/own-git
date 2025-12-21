package snapshots

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type IndexLine struct {
	Fullpath   string
	BlobHash   string
	FileMode   os.FileMode 
	FileSize   int64
	TimeStamps time.Time // For now, still not sure
}

func NewIndexLine() *IndexLine {
	return &IndexLine{}
}

type Staged struct {
	IndexLines []IndexLine
}

func NewStaged() *Staged {
	return &Staged{
		IndexLines: []IndexLine{},
	}
}

func (s *Staged) addIndexLine(indexLine IndexLine) {
	s.IndexLines = append(s.IndexLines, indexLine)
}

func (s *Staged) parseIndexFile(path string) error {
	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	stat, err := fi.Stat()
	stat.Mode()
	if err != nil {
		return err
	}
	if stat.Size() < 1 {
		// TODO: I THINK, I'm messing something here
		return nil
	}
	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		text := scanner.Text()
		line := strings.TrimSpace(text)
		parts := strings.SplitN(line, ":", 5)
		if len(parts) != 5 {
			continue
		}

		fileModeParsedValue, err := strconv.ParseUint(parts[2], 0, 32)
		fileSize, err := strconv.ParseInt(parts[3], 10, 64)
		timestamp, err := time.Parse("2006-01-02 15:04:05", parts[4])

		if err != nil {
			continue
		}

		fileMode := os.FileMode(fileModeParsedValue)
		idxLine := NewIndexLine()
		idxLine.Fullpath = parts[0]
		idxLine.BlobHash = parts[1]
		idxLine.FileMode = fileMode
		idxLine.FileSize = fileSize
		idxLine.TimeStamps = timestamp 

		s.addIndexLine(*idxLine)

	}
	return nil
}

func (s *Staged) visitWorkingDirFiles(basePath string) error{

	return nil
}

func HandleAddCommand() error {
	if len(os.Args[2:]) < 1 {
		slog.Info("hint: Maybe you wanted to say 'git add .'?")
		//TODO: FOR LATER
		slog.Info("hint: Disable this message with 'git config advice.addEmptyPathspec false'")
		return fmt.Errorf("invalid add command")
	}
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	fullpath, ok, err := CheckGitFolderExists(path)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("couldn't found index file")
	}
	s := NewStaged()
	if os.Args[3] == "." {
		fmt.Println("every files")
		if err := s.parseIndexFile(fullpath + ROOTDIR); err != nil {
			return err
		}
		// TODO: NOW VISIT EACH FOLDER AND THEIR FILES TO CHECK

	} else {

	}
	return nil
}
