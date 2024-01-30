package sqlitedriver

import (
	_ "embed"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/hackborn/doc"
	"github.com/hackborn/doc_drivers/drivers/sqlite/gen"
	"github.com/hackborn/doc_drivers/drivers/sqlite/ref"
	"github.com/hackborn/doc_drivers/registry"
	"github.com/hackborn/onefunc/errors"
)

func init() {
	// Register the factory
	const sqlite = "sqlite"
	const driverName = "ref/" + sqlite
	refFn := func() doc.Driver {
		return sqliterefdriver.NewDriver(sqlite)
	}
	genFn := func() doc.Driver {
		return sqlitegendriver.NewDriver(sqlite)
	}
	// Database path is relative to the commands. Relocate it to myself.
	dbpath := filepath.Join("..", "..", "drivers", "sqlite", "data", "db")
	f := registry.Factory{
		Name:       sqlite,
		DriverName: "ref/" + sqlite,
		DbPath:     dbpath,
		ReferenceFiles: map[string]string{
			"const":    refConstGo,
			"driver":   refDriverGo,
			"fn":       refFnGo,
			"metadata": refMetadataGo,
		},
		NewRef:          refFn,
		NewGenerated:    genFn,
		ProcessTemplate: makeTemplates,
	}
	errors.Panic(registry.Register(f))
}

//go:embed ref/ref_const.go
var refConstGo string

//go:embed ref/ref_driver.go
var refDriverGo string

//go:embed ref/ref_fn.go
var refFnGo string

//go:embed ref/ref_metadata.go
var refMetadataGo string