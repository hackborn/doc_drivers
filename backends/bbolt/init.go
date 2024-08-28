package bboltbackend

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/hackborn/doc"
	"github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"

	bboltgendriver "github.com/hackborn/doc_drivers/backends/bbolt/gen"
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
	f.Prepare = newPrepareFunc()

	errors.Panic(registry.Register(f))
}

func addGraphs(m map[string]graphs.Entry) {
	entries, _ := graphs.ReadEntries(graphsFs, "graphs/*.txt")
	for k, v := range entries {
		m[k] = v
	}
}

func newOpenFunc(registry.Factory) func() error {
	return func() error {
		nodes.RegisterNodes()

		// Make drivers accessible to nodes without going through the backend
		refFn := func() doc.Driver {
			inner := bboltrefdriver.NewDriver(nodes.FormatBbolt)
			return &openingDriver{inner: inner}
		}
		genFn := func() doc.Driver {
			inner := bboltgendriver.NewDriver(nodes.FormatBbolt)
			return &openingDriver{inner: inner}
		}

		doc.Register("ref/"+nodes.FormatBbolt, refFn())
		doc.Register("gen/"+nodes.FormatBbolt, genFn())
		return nil
	}
}

func newPrepareFunc() registry.PrepareRunFunc {
	return func(f registry.Factory, graphName string, vars map[string]any) {
	}
}

// openingDriver wraps the actual driver and adds one responsibility,
// to clear out the database file. It works because the Open() function
// returns a new driver. No other functions in this driver should be called.
type openingDriver struct {
	inner doc.Driver
}

func (d *openingDriver) Open(dataSourceName string) (doc.Driver, error) {
	err := os.Remove(dataSourceName)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return d.inner.Open(dataSourceName)
}

func (d *openingDriver) Close() error {
	panic("should not be called")
}

func (d *openingDriver) Format() doc.Format {
	panic("should not be called")
}

func (d *openingDriver) Get(req doc.GetRequest, a doc.Allocator) (*doc.Optional, error) {
	panic("should not be called")
}

func (d *openingDriver) Set(req doc.SetRequestAny, a doc.Allocator) (*doc.Optional, error) {
	panic("should not be called")
}

func (d *openingDriver) Delete(req doc.DeleteRequestAny, a doc.Allocator) (*doc.Optional, error) {
	panic("should not be called")
}

//go:embed graphs/*
var graphsFs embed.FS

//go:embed ref/*
var refFs embed.FS
