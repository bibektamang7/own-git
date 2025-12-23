package snapshots

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanAndAddIndexLines(t *testing.T) {
	r := strings.NewReader("/myowngit/go.mod\tlakfjklsfjkasljflsjafjd\t304623\t242\t1766290356086134418\n")

	s := NewStaged()
	err := s.scanAndAddIndexLines(r)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(s.IndexLines))
	assert.Equal(t, "/myowngit/go.mod", s.IndexLines[0].Fullpath)
}
