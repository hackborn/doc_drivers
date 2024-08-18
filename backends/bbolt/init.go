package bboltbackend

import (
	"embed"
	"path/filepath"

	"github.com/hackborn/doc"
	"github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"

	//	"github.com/hackborn/onefunc/pipeline"

	"github.com/hackborn/doc_drivers/backends/bbolt/nodes"
	bboltrefdriver "github.com/hackborn/doc_drivers/backends/bbolt/ref"
	"github.com/hackborn/doc_drivers/graphs"
	"github.com/hackborn/doc_drivers/registry"
)

func init() {
	// Register the filesystem
	pipeline.RegisterFs("bboltref", refFs)

	// Register the factory
	const bbolt = nodes.FormatBbolt

	// Database path is relative to the commands. Relocate it to myself.
	dbpath := filepath.Join("..", "..", "backends", "bbolt", "data", "db.bbolt")
	graphEntries := graphs.Entries()
	addGraphs(graphEntries)
	f := registry.NewFactory(graphEntries)
	f.Name = bbolt
	f.DbPath = dbpath
	f.Open = newOpenFunc(f)

	errors.Panic(registry.Register(f))
}

func addGraphs(m map[string]graphs.Entry) {
	entries, _ := graphs.ReadEntries(graphsFs, "graphs/*.txt")
	for k, v := range entries {
		m[k] = v
	}
}

func newOpenFunc(f registry.Factory) func() error {
	return func() error {
		nodes.RegisterNodes()

		// Make drivers accessible to nodes without going through the backend
		refFn := func() doc.Driver {
			return bboltrefdriver.NewDriver(nodes.FormatBbolt)
		}
		genFn := func() doc.Driver {
			return bboltrefdriver.NewDriver(nodes.FormatBbolt)
		}

		doc.Register("ref/"+nodes.FormatBbolt, refFn())
		doc.Register("gen/"+nodes.FormatBbolt, genFn())
		return nil
	}
}

//go:embed graphs/*
var graphsFs embed.FS

//go:embed ref/*
var refFs embed.FS
