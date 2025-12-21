package ini

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseINIFile(t *testing.T) {
	data := "[core]\n     defaultBranch = main   \n   bare = false  \n"
	fi := NewFileINI()
	err := fi.ParseINIFile(strings.NewReader(data))

	assert.NoError(t, err)
	assert.Equal(t, 3, len(fi.lines))
}

func TestIniAdd(t *testing.T) {
	fi := NewFileINI()
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "defaultBranch", "main")

	assert.Equal(t, 3, len(fi.lines))
}

func TestINISet(t *testing.T) {
	fi := NewFileINI()
	fi.Add("core", "defaultBranch", "main")
	// fi.Add("core", "defaultBranch", "main")
	ok := fi.Set("core", "defaultBranch", "master")

	assert.Equal(t, ok, true)
	assert.Equal(t, "master", fi.lines[1].Value)
}

func TestINIUnset(t *testing.T) {
	fi := NewFileINI()
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "bare", "false")
	// fi.Add("core", "defaultBranch", "main")
	ok := fi.Unset("core", "defaultBranch")

	assert.Equal(t, ok, true)
	assert.Equal(t, 2, len(fi.lines))
}

func TestINIUnsetAll(t *testing.T) {
	fi := NewFileINI()
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "bare", "false")
	fi.Add("core", "defaultBranch", "main")
	fi.UnsetAll("core", "defaultBranch")

	assert.Equal(t, 2, len(fi.lines))

}

func TestINIReplaceAll(t *testing.T) {

	fi := NewFileINI()
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "bare", "false")
	fi.Add("core", "defaultBranch", "main")
	fi.ReplaceAll("core", "defaultBranch", "master")

	assert.Equal(t, 3, len(fi.lines))

	newFi := NewFileINI()
	newFi.Add("core", "defaultBranch", "main")
	newFi.Add("core", "bare", "false")
	newFi.ReplaceAll("core", "defaultBranch", "master")

	assert.Equal(t, 3, len(fi.lines))

}

func TestINIRemoveSection(t *testing.T) {
	fi := NewFileINI()
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "defaultBranch", "main")
	fi.Add("core", "bare", "false")
	fi.Add("core", "defaultBranch", "main")
	fi.Add("user", "name", "bibek")
	fi.RemoveSection("core")

	assert.Equal(t, 2, len(fi.lines))

	fi.Add("core", "defaultBranch", "main")
	fi.Add("user", "name", "bibek")
	fi.Add("remote", "branch", "main")
	fi.RemoveSection("user")

	assert.Equal(t, 4, len(fi.lines))


}
