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
	const sqlite = nodes.FormatSqlite

	// Database path is relative to the commands. Relocate it to myself.
	dbpath := filepath.Join("..", "..", "backends", "sqlite", "data", "db")
	graphEntries := graphs.Entries()
	addGraphs(graphEntries)
	f := registry.NewFactory(graphEntries)
	f.Name = sqlite
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
			return sqliterefdriver.NewDriver(nodes.FormatSqlite)
		}
		genFn := func() doc.Driver {
			return sqlitegendriver.NewDriver(nodes.FormatSqlite)
		}
		doc.Register("ref/"+nodes.FormatSqlite, refFn())
		doc.Register("gen/"+nodes.FormatSqlite, genFn())

		return nil
	}
}

//go:embed graphs/*
var graphsFs embed.FS

//go:embed ref/*
var refFs embed.FS

//go:embed templates/*
var templatesFs embed.FS
