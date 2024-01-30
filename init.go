package drivers

import (
	"embed"
	"log"
	"path/filepath"
	"slices"
	"strings"
)

func init() {
	entries, err := graphs.ReadDir("graphs")
	if err != nil {
		log.Fatal("init err: ", err)
	}

	for _, entry := range entries {
		name := entry.Name()
		ext := filepath.Ext(name)
		if strings.ToLower(ext) == ".txt" {
			name = strings.TrimSuffix(name, ext)
			graphNames = append(graphNames, name)
			graphEntries[name] = filepath.Join("graphs", entry.Name())
		}
	}
	slices.Sort(graphNames)
}

//go:embed graphs/*
var graphs embed.FS
