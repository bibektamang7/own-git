package snapshots

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type IndexLine struct {
	Fullpath   string
	BlobHash   string
	FileMode   uint32
	FileSize   int64
	TimeStamps int64
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

func getGitMode(mode os.FileMode) uint32 {
	if mode.IsDir() {
		return 040000 // Git directory mode
	}
	// Check if any execute bit is set
	if mode&0111 != 0 {
		return 100755 // Git executable
	}
	return 100644 // Git regular file
}

func (s *Staged) parseIndexFile(path string) error {
	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	stat, err := fi.Stat()
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
		parts := strings.SplitN(line, "\t", 5)
		fmt.Println("lenght of line: ", len(parts))
		if len(parts) != 5 {
			continue
		}

		fileModeParsedValue, err := strconv.ParseUint(parts[2], 0, 32)
		fileSize, err := strconv.ParseInt(parts[3], 10, 64)
		timestamp, err := strconv.ParseInt(parts[4], 10, 64)

		if err != nil {
			fmt.Println("here comes", err)
			continue
		}

		fileMode := os.FileMode(fileModeParsedValue)
		idxLine := IndexLine{
			Fullpath:   parts[0],
			BlobHash:   parts[1],
			FileMode:   getGitMode(fileMode),
			FileSize:   fileSize,
			TimeStamps: timestamp,
		}

		s.addIndexLine(idxLine)

	}
	fmt.Println("length of parse index file: ", len(s.IndexLines))
	return nil
}

func (s *Staged) addFileInfoInIndexLines(path string, entry os.DirEntry, currentIdx int) error {

	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fi.Close()

	info, err := entry.Info()
	if err != nil {
		return err
	}
	h := sha1.New()
	header := fmt.Sprintf("blob %d\x00", info.Size())
	h.Write([]byte(header))

	io.Copy(h, fi)
	fileHash := hex.EncodeToString(h.Sum(nil))

	fmt.Println("the hash: ", fileHash)

	idxLine := IndexLine{
		Fullpath:   path,
		FileMode:   getGitMode(info.Mode()),
		FileSize:   info.Size(),
		TimeStamps: info.ModTime().UnixNano(),
		BlobHash:   fileHash,
	}

	s.addIndexLine(idxLine)

	return nil
}

func (s *Staged) compareAndAddToIndex(path string, entry os.DirEntry, currentIdx int) error {
	return nil
}

func (s *Staged) visitWorkingDirFilesAndCompare(basePath string) error {
	stack := []string{basePath}
	isComparable := len(s.IndexLines) > 0
	if isComparable {
		for _, line := range s.IndexLines {
			fmt.Printf("%s\t%s\t%o\t%d\t%d\n",
				line.Fullpath,
				line.BlobHash,
				line.FileMode,
				line.FileSize,
				line.TimeStamps,
			)
		}
	}
	for len(stack) > 0 {
		currentDir := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		entries, err := os.ReadDir(currentDir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			path := filepath.Join(currentDir, entry.Name())

			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == ".owngit" {
					continue
				}
				stack = append(stack, path)
			} else {
				if isComparable {

				} else {
					s.addFileInfoInIndexLines(path, entry, 0)
				}
				fmt.Println("Processing file:", path)
			}
		}
	}
	return nil
}

func (s *Staged) addIndexLineToIndexFile(path string) error {
	lockIndex := "index.lock"

	fi, err := os.Create(path + lockIndex)
	if err != nil {
		return err
	}
	defer fi.Close()
	bufWriter := bufio.NewWriter(fi)

	for _, line := range s.IndexLines {

		fmt.Fprintf(bufWriter, fmt.Sprintf("%s\t%s\t%o\t%d\t%d\n",
			line.Fullpath,
			line.BlobHash,
			line.FileMode,
			line.FileSize,
			line.TimeStamps,
		))
	}
	if err := bufWriter.Flush(); err != nil {
		return err
	}
	fi.Close()
	return os.Rename(path+lockIndex, path+"index")
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
	if os.Args[2] == "." {
		fmt.Println("every files")
		if err := s.parseIndexFile(fullpath + ROOTDIR + "index"); err != nil {
			return err
		}
		if err := s.visitWorkingDirFilesAndCompare(path); err != nil {
			return err
		}

	} else {

		fmt.Println("for now, nothing")
	}
	if err := s.addIndexLineToIndexFile(fullpath + ROOTDIR); err != nil {
		return err
	}
	return nil
}
