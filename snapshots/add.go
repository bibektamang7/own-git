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
		return 040000
	}
	if mode&0111 != 0 {
		return 100755
	}
	return 100644
}

func (s *Staged) scanAndAddIndexLines(r io.Reader) error {
	scanner := bufio.NewScanner(r)
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
			return err
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
	return nil
}

func (s *Staged) parseIndexFile(path string) error {
	fi, err := os.Open(path)
	if err != nil {
		return err
	}
	if err := s.scanAndAddIndexLines(fi); err != nil {
		return err
	}
	fmt.Println("length of parse index file: ", len(s.IndexLines))
	return nil
}

func getIndexLine(path string, entry os.DirEntry) (IndexLine, error) {
	fi, err := os.Open(path)
	if err != nil {
		return IndexLine{}, err
	}
	defer fi.Close()

	info, err := entry.Info()
	if err != nil {
		return IndexLine{}, err
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
	return idxLine, nil

}

func (s *Staged) addFileInfoInIndexLines(path string, entry os.DirEntry, currentIdx int) error {
	if currentIdx >= len(s.IndexLines) {
		fmt.Println("hello")
		return fmt.Errorf("invalid current index")
	}
	idxLine, err := getIndexLine(path, entry)
	if err != nil {
		return err
	}

	s.IndexLines[currentIdx] = idxLine
	return nil
}

func (s *Staged) visitWorkingDirFiles(basePath string) error {
	stack := []string{basePath}
	idxIndexLinesCount := 0
	idxIndexLinesSize := len(s.IndexLines)
	isComparable := idxIndexLinesSize > 0
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
					if idxIndexLinesCount+1 >= len(s.IndexLines) {
						isComparable = false
					}
				outer:
					for i := idxIndexLinesCount; i < len(s.IndexLines); i++ {
						fmt.Println("fullpath: ", s.IndexLines[i].Fullpath)
						fmt.Println("path: ", path)
						if s.IndexLines[i].Fullpath == path {
							fmt.Println("how many time equal: ")
							info, err := entry.Info()
							if err != nil {
								return err
							}
							if s.IndexLines[i].FileSize != info.Size() && s.IndexLines[i].TimeStamps != info.ModTime().UnixNano() {
								fmt.Println("does it change")
								idxLine, err := getIndexLine(path, entry)
								if err != nil {
									return err
								}
								if s.IndexLines[i].BlobHash != idxLine.BlobHash {
									s.IndexLines[i] = idxLine
								}
							}
							idxIndexLinesCount++
							break outer

						} else if s.IndexLines[i].Fullpath < path && len(s.IndexLines[i].Fullpath) <= len(path) {
							fmt.Println("how many time less than path: ")
							//FILE DELETED CASE
							if strings.HasSuffix(s.IndexLines[i].Fullpath, currentDir+entry.Name()) {
								// s.addFileInfoInIndexLines(path, entry, idxIndexLinesCount)
								fmt.Println("is it here when not to")
								continue
							}
						} else {
							if len(s.IndexLines[i].Fullpath) < len(path) {
								continue
							}
							fmt.Println("how many time greater than path: ")
							// NEW FILE
							idxLine, err := getIndexLine(path, entry)
							if err != nil {
								return err
							}
							s.IndexLines = append(s.IndexLines[:i], append([]IndexLine{idxLine}, s.IndexLines[i:]...)...)
							idxIndexLinesCount++
							break outer
						}

						if idxIndexLinesCount == 4 {
							panic("let's see")
						}
						idxIndexLinesCount++
					}
				} else {
					idxLine, err := getIndexLine(path, entry)
					if err != nil {
						return err
					}
					s.IndexLines = append(s.IndexLines, idxLine)
					idxIndexLinesCount++
				}
			}
		}
	}
	s.IndexLines = s.IndexLines[:idxIndexLinesCount]
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
	fmt.Println("full path in handle command:  ", fullpath)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("couldn't found index file %v", err)
	}
	s := NewStaged()
	if os.Args[2] == "." {
		if err := s.parseIndexFile(fullpath + ROOTDIR + "index"); err != nil {
			return err
		}
		if err := s.visitWorkingDirFiles(path); err != nil {
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
