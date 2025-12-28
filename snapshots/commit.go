package snapshots

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ContentType string

const (
	Blob ContentType = "blob"
	Tree ContentType = "tree"
)

type CommitTree struct {
	fileMode    string
	contentType ContentType
	Hash        string
	Name        string
}

type Commit struct {
	tree     string
	parent   string
	author   []string
	commiter []string
	message  string
}

type TreePaths struct {
	TreePaths map[string]string
}

var (
	ERROR_MALFORMED_TREE_FORMAT   = fmt.Errorf("malformed tree format")
	ERROR_MALFORMED_COMMIT_FORMAT = fmt.Errorf("malformed commit format")
)

func NewTreePaths() TreePaths {
	return TreePaths{
		TreePaths: make(map[string]string),
	}
}

func (t *TreePaths) parseTreeFile(r io.Reader, gitRoot, treePath string) error {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) != 4 {
			return ERROR_MALFORMED_TREE_FORMAT
		}
		filePath := treePath + parts[3]
		if parts[1] == "tree" {
			treeHash := strings.TrimSpace(parts[2])
			treeHashParts := []string{treeHash[:2], treeHash[2:]}
			if len(treeHashParts) != 2 {
				return ERROR_MALFORMED_COMMIT_FORMAT
			}
			// treeHashFilePath := gitRoot + ROOTDIR + "objects/" + treeHashParts[0] + "/" + treeHashParts[1]
			// TODO: IF IT WORKS THEN CHANGE OTHER TOO
			treeHashFilePath := filepath.Join(gitRoot, ROOTDIR, "objects", treeHashParts[0], treeHashParts[1])
			treeHashFile, err := os.Open(treeHashFilePath)
			if err != nil {
				return err
			}

			if err := t.parseTreeFile(treeHashFile, gitRoot, filePath); err != nil {
				return err
			}

			treeHashFile.Close()
			continue
		}

		// assigning blob to tree path, which is useful
		// when comparing with index lines
		t.TreePaths[filePath] = parts[2]
	}
}

func parseCommitFile(r io.Reader, basePath string) (TreePaths, error) {
	treeReader := bufio.NewReader(r)

	treeLine, err := treeReader.ReadString('\n')
	if err != nil {
		return TreePaths{}, err
	}

	treeParts := strings.SplitN(treeLine, " ", 2)
	if len(treeParts) != 2 {
		return TreePaths{}, ERROR_MALFORMED_COMMIT_FORMAT
	}
	if treeParts[0] != "tree" {
		return TreePaths{}, ERROR_MALFORMED_COMMIT_FORMAT
	}
	treeHash := strings.TrimSpace(treeParts[1])
	treeHashParts := []string{treeHash[:2], treeHash[2:]}
	if len(treeHashParts) != 2 {
		return TreePaths{}, ERROR_MALFORMED_COMMIT_FORMAT
	}
	treeHashFilePath := basePath + ROOTDIR + "objects/" + treeHashParts[0] + "/" + treeHashParts[1]
	treeHashFile, err := os.Open(treeHashFilePath)
	if err != nil {
		return TreePaths{}, err
	}
	defer treeHashFile.Close()

	treePaths := NewTreePaths()
	if err := treePaths.parseTreeFile(treeHashFile, basePath, ""); err != nil {
		return TreePaths{}, err
	}
	return treePaths, nil
}

func ParseHeadFile(basePath string) (TreePaths, error) {
	fi, err := os.Open(basePath + ROOTDIR + "HEAD")
	if err != nil {
		return TreePaths{}, err
	}
	defer fi.Close()

	reader := bufio.NewReader(fi)

	commitHash, err := reader.ReadString('\n')
	if err != nil {
		return TreePaths{}, err
	}

	commitTrimHash := strings.TrimSpace(commitHash)
	parts := []string{commitTrimHash[:2], commitTrimHash[2:]}
	if len(parts) != 2 {
		return TreePaths{}, ERROR_MALFORMED_COMMIT_FORMAT
	}
	commitFilePath := basePath + ROOTDIR + "objects/" + parts[0] + "/" + parts[1]
	commitFile, err := os.Open(commitFilePath)
	if err != nil {
		return TreePaths{}, err
	}
	defer commitFile.Close()
	return parseCommitFile(commitFile, basePath)
}
func groupIndexByDir(index []IndexLine) map[string][]IndexLine {
	dirs := make(map[string][]IndexLine)

	for _, line := range index {
		dir := filepath.Dir(line.Fullpath)
		if dir == "." {
			dir = ""
		}
		dirs[dir] = append(dirs[dir], line)
	}
	return dirs
}
func sortedDirsByDepth(dirs map[string][]IndexLine) []string {
	keys := make([]string, 0, len(dirs))
	for k := range dirs {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return strings.Count(keys[i], "/") > strings.Count(keys[j], "/")
	})
	return keys
}
func writeTreeObject(
	gitRoot string,
	entries []CommitTree,
) (string, error) {

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	var buf strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&buf, "%s\t%s\t%s\t%s\n",
			e.fileMode,
			e.contentType,
			e.Hash,
			e.Name,
		)
	}

	content := buf.String()
	header := fmt.Sprintf("tree %d\x00", len(content))
	hash := hashBytes([]byte(header + content))

	objPath := objectPath(gitRoot, hash)
	if err := writeObject(objPath, content); err != nil {
		return "", err
	}

	return hash, nil
}
func writeObject(path string, content string) error {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// If object already exists, do nothing (Git behavior)
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func objectPath(gitRoot, hash string) string {
	dir := hash[:2]
	file := hash[2:]
	return filepath.Join(gitRoot, ROOTDIR, "objects", dir, file)
}

func hashBytes(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func buildTreesFromIndex(
	gitRoot string,
	index []IndexLine,
) (string, error) {
	dirMap := groupIndexByDir(index)
	treeHashes := make(map[string]string)

	dirs := sortedDirsByDepth(dirMap)

	for _, dir := range dirs {
		var entries []CommitTree

		// files in this directory
		for _, line := range dirMap[dir] {
			entries = append(entries, CommitTree{
				fileMode:    fmt.Sprintf("%o", line.FileMode),
				Name:        filepath.Base(line.Fullpath),
				Hash:        line.BlobHash,
				contentType: Blob,
			})
		}

		// child trees
		for childDir, childHash := range treeHashes {
			if filepath.Dir(childDir) == dir {
				entries = append(entries, CommitTree{
					fileMode:    "40000",
					Name:        filepath.Base(childDir),
					Hash:        childHash,
					contentType: Tree,
				})
			}
		}

		treeHash, err := writeTreeObject(gitRoot, entries)
		if err != nil {
			return "", err
		}

		treeHashes[dir] = treeHash
	}

	return treeHashes[""], nil 
}

func compareAndFindStagedFiles(gitRootPath string) error {
	staged := NewStaged()

	staged.baseRoot = gitRootPath
	if err := staged.parseIndexFile(); err != nil {
		return err
	}
	treePaths, err := ParseHeadFile(gitRootPath)
	if err != nil {
		return err
	}
	changed := false

	for path, idx := range staged.indexMap {
		idxLine := staged.IndexLines[idx]

		if headHash, ok := treePaths.TreePaths[path]; !ok {
			changed = true // new file
			break
		} else if headHash != idxLine.BlobHash {
			changed = true // modified
			break
		}
	}

	// deleted files
	if !changed {
		for path := range treePaths.TreePaths {
			if _, ok := staged.indexMap[path]; !ok {
				changed = true
				break
			}
		}
	}

	if !changed {
		fmt.Println("On branch main")
		fmt.Println("nothing to commit, working tree clean")
		return nil
	}

	// TODO: NOW CREATE COMMIT FORMAT AND TREE
	treeHash, err := buildTreesFromIndex(gitRootPath, staged.IndexLines)
	if err != nil {
		return err
	}
	return nil
}

func HandleCommitCommand() error {
	fs := flag.NewFlagSet("commit", flag.ExitOnError)
	msg := fs.String("m", "", "commit message")

	fs.Parse(os.Args[2:])
	args := fs.Args()

	if len(args) > 0 {
		log.Fatalf("invalid command argument: %s\n", args[0])
	}
	if len(*msg) < 1 {
		return fmt.Errorf("empty commit message")
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Println("message: ", *msg)

	_, ok, err := CheckGitFolderExists(path)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("outside of Git repository")
	}

	return nil
}
