package ini

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// INI file formet
// It is a configuration file for computer software that
// consists of plain text with a structure and syntax comparising
// key-value parirs organized in sections.

type KeyValue map[string]string
type INI map[string]KeyValue

type Raw string

type Line struct {
	Section string
	Raw     string
	Key     string
	Value   string
	IsSec   bool
	IsKV    bool
}

type FileINI struct {
	lines []Line
}

func NewFileINI() *FileINI {
	return &FileINI{
		lines: []Line{},
	}
}

func (fi *FileINI) ParseINIFile(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	var lines []Line
	section := ""

	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		l := Line{Raw: raw, IsSec: false, IsKV: false}

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			lines = append(lines, l)
			continue
		}

		if strings.HasPrefix(line, "[") || strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			l.IsSec = true
			l.Section = section
			lines = append(lines, l)
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			l.Section = section
			l.Key = key
			l.Value = value
			l.IsKV = true
		}

		lines = append(lines, l)

	}
	return nil
}

func (fi *FileINI) Add(section, key, value string) {
	newLine := Line{
		Section: section,
		Key:     key,
		Value:   value,
		IsKV:    true,
		IsSec:   false,
	}

	for i := len(fi.lines) - 1; i >= 0; i-- {
		if fi.lines[i].IsSec && fi.lines[i].Section == section {
			fi.lines = append(
				fi.lines[:i+1],
				append([]Line{newLine}, fi.lines[i+1:]...)...,
			)
			return
		}

	}

	fi.lines = append(
		fi.lines,
		Line{Raw: fmt.Sprintf("[%s]", section), Section: section, IsKV: false, IsSec: true},
		newLine,
	)

}

func (fi *FileINI) Get(section, key string) {}

func (fi *FileINI) Set(section, key, value string) {
}

func (fi *FileINI) Unset() {
}

func (fi *FileINI) Replace(section, key, value string) {
}

func (fi *FileINI) Write(w io.Writer) error {
	return nil
}
