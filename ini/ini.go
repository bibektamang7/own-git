package ini

// INI file formet
// It is a configuration file for computer software that
// consists of plain text with a structure and syntax comparising
// key-value parirs organized in sections.

type INI map[string]map[string]string

func ReadINIFile() error {
	return nil
}
