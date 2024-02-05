package sqlitebackend

import (
	"embed"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/hackborn/doc"
	"github.com/hackborn/doc_drivers/backends/sqlite/gen"
	"github.com/hackborn/doc_drivers/backends/sqlite/nodes"
	"github.com/hackborn/doc_drivers/backends/sqlite/ref"
	"github.com/hackborn/doc_drivers/graphs"
	"github.com/hackborn/doc_drivers/registry"
	"github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
)

func init() {
	// Register the filesystem
	pipeline.RegisterFs("sqliteref", refFs)
	pipeline.RegisterFs(nodes.TemplateFsName, templatesFs)

	// Register the factory
	const sqlite = "sqlite"
	const driverName = "ref/" + sqlite
	refFn := func() doc.Driver {
		return sqliterefdriver.NewDriver(sqlite)
	}
	genFn := func() doc.Driver {
		return sqlitegendriver.NewDriver(sqlite)
	}
	// Make this accessible to nodes without going through this driver.
	// This raises the question of why I have these factory functions at all.
	// Unless I see a downside to it, I'll probably remove them.
	doc.Register(driverName, refFn())
	doc.Register("gen/"+sqlite, genFn())

	// Database path is relative to the commands. Relocate it to myself.
	dbpath := filepath.Join("..", "..", "backends", "sqlite", "data", "db")
	graphEntries := graphs.Entries()
	addGraphs(graphEntries)
	f := registry.NewFactory(graphEntries)
	f.Name = sqlite
	f.DriverName = driverName
	f.DbPath = dbpath
	f.NewRef = refFn
	f.NewGenerated = genFn
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
		return nil
	}
}

//go:embed graphs/*
var graphsFs embed.FS

//go:embed ref/*
var refFs embed.FS

//go:embed templates/*
var templatesFs embed.FS
