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
	"sort"
	"strconv"
	"strings"
)

// IndexLine represents each line from Index file
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
	indexMap   map[string]int
	seen       map[string]bool
	currentDir string
	baseRoot   string
}

func NewStaged() *Staged {
	return &Staged{
		IndexLines: []IndexLine{},
		indexMap:   make(map[string]int),
		seen:       make(map[string]bool),
		currentDir: "",
		baseRoot:   "",
	}
}
func getGitMode(mode os.FileMode) uint32 {
	if mode&os.ModeSymlink != 0 {
		return 120000
	}
	if mode&0111 != 0 {
		return 100755
	}
	return 100644
}

func hashFile(path string, info os.FileInfo) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	header := fmt.Sprintf("blob %d\x00", info.Size())
	h.Write([]byte(header))

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func (s *Staged) parseIndexFile() error {
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

		s.indexMap[idx.Fullpath] = len(s.IndexLines)
		s.IndexLines = append(s.IndexLines, idx)
	}

	return nil
}
func (s *Staged) visitWorkingDirFiles(repoRoot string) error {
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

		if idx, ok := s.indexMap[rel]; ok {
			old := s.IndexLines[idx]

			if old.FileSize != info.Size() ||
				old.TimeStamps != info.ModTime().UnixNano() {

				hash, err := hashFile(path, info)
				if err != nil {
					return err
				}

				if hash != old.BlobHash {
					s.IndexLines[idx] = IndexLine{
						Fullpath:   rel,
						BlobHash:   hash,
						FileMode:   getGitMode(info.Mode()),
						FileSize:   info.Size(),
						TimeStamps: info.ModTime().UnixNano(),
					}
				}
			}
			return nil
		}

		// New file
		hash, err := hashFile(path, info)
		if err != nil {
			return err
		}

		s.indexMap[rel] = len(s.IndexLines)
		s.IndexLines = append(s.IndexLines, IndexLine{
			Fullpath:   rel,
			BlobHash:   hash,
			FileMode:   getGitMode(info.Mode()),
			FileSize:   info.Size(),
			TimeStamps: info.ModTime().UnixNano(),
		})

		return nil
	})
}

func (s *Staged) removeDeleted() {
	filtered := s.IndexLines[:0]

	for _, line := range s.IndexLines {
		if s.currentDir != "" && !strings.HasPrefix(line.Fullpath, s.currentDir+"/") && line.Fullpath != s.currentDir {
			filtered = append(filtered, line)
			continue
		}
		if s.seen[line.Fullpath] {
			filtered = append(filtered, line)
		}
	}
	s.IndexLines = filtered
}

func getAddRoot(repoRoot, cwd string) (string, error) {
	rel, err := filepath.Rel(repoRoot, cwd)
	if err != nil {
		return "", err
	}
	if rel == "." {
		return "", nil
	}
	return filepath.ToSlash(rel), nil

}

func (s *Staged) writeIndex(path string) error {
	sort.Slice(s.IndexLines, func(i, j int) bool {
		return s.IndexLines[i].Fullpath < s.IndexLines[j].Fullpath
	})

	lock := path + "index.lock"
	f, err := os.Create(lock)
	if err != nil {
		return err
	}

	w := bufio.NewWriter(f)
	for _, line := range s.IndexLines {
		fmt.Fprintf(w, "%s\t%s\t%o\t%d\t%d\n",
			line.Fullpath,
			line.BlobHash,
			line.FileMode,
			line.FileSize,
			line.TimeStamps,
		)
	}

	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	f.Close()

	return os.Rename(lock, path+"index")
}

func (s *Staged) addSpecificFiles(path string) error {
	convertedPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	currentRelPath, err := filepath.Rel(s.baseRoot, convertedPath)
	if err != nil {
		return err
	}
	if strings.HasPrefix(currentRelPath, "..") {
		return fmt.Errorf("%s is outside repository at %s\n", currentRelPath, s.baseRoot)
	}
	fi, err := os.Open(convertedPath)
	if err != nil {
		return err
	}
	defer fi.Close()
	if line, ok := s.indexMap[currentRelPath]; ok {
		fmt.Println("the line : ", line, "the is rel : ", currentRelPath)
		return nil
	}
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
		return fmt.Errorf("couldn't found index file %v", err)
	}
	s := NewStaged()

	s.baseRoot = fullpath
	root, err := getAddRoot(fullpath, path)
	if err != nil {
		return err
	}
	s.currentDir = root
	if err := s.parseIndexFile(); err != nil {
		return err
	}
	if os.Args[2] == "." {
		if err := s.visitWorkingDirFiles(path); err != nil {
			return err
		}
		s.removeDeleted()
	} else {
		for _, p := range os.Args[2:] {
			if err := s.addSpecificFiles(p); err != nil {
				return err
			}
		}
	}
	return s.writeIndex(fullpath + ROOTDIR)
}
