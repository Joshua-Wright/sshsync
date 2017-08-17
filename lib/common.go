package ssh_sync

import (
	"strings"
	"os"
	"log"
)

var endings []string = []string{
	".cpp",
	".hpp",
	".c",
	".h",
	".go",
	".hs",
	".cl",
	".js",
	".md",
	".txt",
}

var ignored_prefixes = []string{
	".git",
	".realtime",
	".idea",
}

type IgnoreConfig struct {
	/*TODO*/
	_unexported int
}

func (c *IgnoreConfig) ShouldIgnore(path string) bool {
	for _, prefix := range ignored_prefixes {
		if strings.HasPrefix(path, prefix) {
			//log.Println("ignoring ", path)
			return true
		}
	}

	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		// skip checking endings on directories
		return false
	} else if err != nil {
		// ignoreConfig things we can't stat
		return true
	}

	for _, ending := range endings {
		if strings.HasSuffix(path, ending) {
			//log.Println("adding ", path)
			return false
		}
	}
	return true
}

func logFatalIfNotNil(label string, err error) {
	if err != nil {
		log.Fatal(label, " error: ", err)
	}
}