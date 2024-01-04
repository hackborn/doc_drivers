package sqlitedriver

import (
	_ "embed"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/hackborn/doc"
	"github.com/hackborn/doc_drivers/registry"
	"github.com/hackborn/onefunc/errors"
)

func init() {
	// Register the factory
	const sqlite = "sqlite"
	const driverName = "ref/" + sqlite
	fn := func() doc.Driver {
		return &_toxDriver{sqlDriverName: sqlite}
	}
	// Database path is relative to the commands. Relocate it to myself.
	dbpath := filepath.Join("..", "..", "sqlite", "data", "db")
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
		New:             fn,
		ProcessTemplate: makeTemplates,
	}
	errors.Panic(registry.Register(f))
}

//go:embed ref_const.go
var refConstGo string

//go:embed ref_driver.go
var refDriverGo string

//go:embed ref_fn.go
var refFnGo string

//go:embed ref_metadata.go
var refMetadataGo string
