package ini

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// INI file formet
// It is a configuration file for computer software that
// consists of plain text with a structure and syntax comparising
// key-value parirs organized in sections.

// represents INI file line
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

// loads/parses INI file lines
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

// adds key with value into section
func (fi *FileINI) Add(section, key, value string) {
	newLine := Line{
		Section: section,
		Key:     key,
		Value:   value,
		IsKV:    true,
		IsSec:   false,
		Raw:     fmt.Sprintf("\t\t%s = %s", key, value),
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

func (fi *FileINI) Get(section, key string) string {
	for _, line := range fi.lines {
		if line.IsKV && line.Section == section && line.Key == key {
			return line.Value
		}
	}
	return ""
}

func (fi *FileINI) GetAll(section, key string) []string {
	values := []string{}
	for _, line := range fi.lines {
		if line.IsKV && line.Section == section && line.Key == key {
			values = append(values, line.Value)
		}
	}
	return values
}

// replaces existing key's value from section
func (fi *FileINI) Set(section, key, value string) bool {
	count := 0
	idx := -1
	for i, line := range fi.lines {
		if line.IsKV && line.Section == section && line.Key == key {
			count++
			idx = i
			if count > 1 {
				slog.Warn(fmt.Sprintf("%s.%s has multiple values", section, key))
				slog.Error("Cannot overwrite multiple values with a single value")
				slog.Info("Use a regexp, --add or --replace-all to change user.name")
				return false
			}
		}
	}
	if idx == -1 {
		return false
	}
	if fi.lines[idx].Value == value {
		return false
	}

	fi.lines[idx].Value = value

	return true
}

// Unset key from the section
// Only unique / non-duplicate key
func (fi *FileINI) Unset(section, key string) bool {
	count := 0
	idx := -1
	for i, line := range fi.lines {
		if line.IsKV && line.Section == section && line.Key == key {
			count++
			idx = i
			if count > 1 {
				slog.Warn(fmt.Sprintf("%s.%s has multiple values", section, key))
				slog.Error("Cannot overwrite multiple values with a single value")
				slog.Info("Use a regexp, --add or --replace-all to change user.name")
				return false
			}
		}
	}
	if idx == -1 {
		return false
	}
	removeIdxCount := 0
	if fi.lines[idx-1].IsSec {
		removeIdxCount = 1
	}
	fi.lines = append(fi.lines[:idx-removeIdxCount], fi.lines[idx+1:]...)
	return true

}

// Unsets duplicate key from the section
func (fi *FileINI) UnsetAll(section, key string) {

	count := 0

	for i, line := range fi.lines {
		if line.IsKV && line.Section == section && line.Key == key {
			continue
		}

		isGivenSection := line.IsSec && line.Section == section
		if isGivenSection {
			isLastLine := i+1 == len(fi.lines)
			isEmptySection := isLastLine || fi.lines[i+1].IsSec
			if isEmptySection {
				continue
			}
		}

		fi.lines[count] = line
		count++
	}
	for k := count; k < len(fi.lines); k++ {
		fi.lines[k] = Line{}
	}
	fi.lines = fi.lines[:count]
}

// replaces duplicate keys to one with new value
func (fi *FileINI) ReplaceAll(section, key, value string) {
	isReplaced := false
	count := 0
	for _, line := range fi.lines {
		if line.IsKV && line.Section == section && line.Key == key {
			if !isReplaced {
				line.Value = value
				fi.lines[count] = line
				count++
				isReplaced = true
			}
			continue
		}
		fi.lines[count] = line
		count++
	}
	for k := count; k < len(fi.lines); k++ {
		fi.lines[k] = Line{}
	}
	fi.lines = fi.lines[:count]
}

func (fi *FileINI) RenameSection(newSection, oldSection string) {
	for i, line := range fi.lines {
		if line.IsSec && line.Section == oldSection {
			fi.lines[i].Section = newSection
			return
		}
	}
}

func (fi *FileINI) RemoveSection(section string) {
	count := 0
	isInTargetSection := false

	for _, line := range fi.lines {
		if line.IsSec {
			if line.Section == section {
				isInTargetSection = true
			} else {
				isInTargetSection = false
			}
		}
		if isInTargetSection {
			continue
		}
		fi.lines[count] = line
		count++
	}

	for k := count; k < len(fi.lines); k++ {
		fi.lines[k] = Line{}
	}

	fi.lines = fi.lines[:count]
}

func (fi *FileINI) List(section string) {
	for _, line := range fi.lines {
		if line.IsKV {
			fmt.Printf("%s.%s = %s\n", line.Section, line.Key, line.Value)
		}
	}
}

func (fi *FileINI) Write(w io.Writer) error {
	bufWriter := bufio.NewWriter(w)

	for _, line := range fi.lines {
		if _, err := fmt.Fprintln(bufWriter, line.Raw); err != nil {
			return err
		}
	}
	return bufWriter.Flush()
}
