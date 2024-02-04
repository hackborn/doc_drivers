package graphs

import (
	"embed"
	"log"
)

func init() {
	initEntries()
}

// initEntries makes an entry for every graph file in resources/.
func initEntries() {
	newEntries, err := ReadEntries(resourcesFs, "resources/*.txt")
	if err != nil {
		log.Fatal("init err: ", err)
	}
	entries = newEntries
}

//go:embed resources/*
var resourcesFs embed.FS
