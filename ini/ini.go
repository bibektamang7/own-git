package ini

import (
	"fmt"
	"io"
)

// INI file formet
// It is a configuration file for computer software that
// consists of plain text with a structure and syntax comparising
// key-value parirs organized in sections.

type KeyValue map[string]string
type INI map[string]KeyValue

type Raw string

func ReadINIFile() error {
	return nil
}
func writeKeyValueINI(w io.WriteCloser, data KeyValue) (int, error) {
	defer w.Close()
	// TODO: THIS ONE IS TRICKY
	return 0, nil
}

func writeRawINI(w io.WriteCloser, data Raw) (int, error) {
	defer w.Close()
	// TODO : THINGS TO DO
	return w.Write([]byte(data))
}

func WriteINI(w io.WriteCloser, content any) (int, error) {
	switch data := content.(type) {
	case KeyValue:
		return writeKeyValueINI(w, data)
	case Raw:
		return writeRawINI(w, data)
	default:
		return 0, fmt.Errorf("invalid content")
	}
}
