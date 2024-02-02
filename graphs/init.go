package graphs

import (
	"embed"
	"log"
	"path/filepath"
	"strings"
)

func init() {
	initEntries()
}

// initEntries makes an entry for every graph file in resources/.
func initEntries() {
	direntries, err := resourcesFs.ReadDir("resources")
	if err != nil {
		log.Fatal("init err: ", err)
	}

	for _, entry := range direntries {
		name := entry.Name()
		ext := filepath.Ext(name)
		if strings.ToLower(ext) == ".txt" {
			name = strings.TrimSuffix(name, ext)
			path := filepath.Join("resources", entry.Name())
			entries[name] = Entry{Graph: NewReadFileFunc(path, resourcesFs)}
		}
	}
}

//go:embed resources/*
var resourcesFs embed.FS
