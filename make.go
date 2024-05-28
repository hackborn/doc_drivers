package drivers

import (
	"github.com/hackborn/doc_drivers/registry"
	"github.com/hackborn/onefunc/pipeline"
)

// MakeDriver generates a new driver with the supplied settings.
func MakeDriver(settings MakeDriverSettings) error {
	f, err := registry.Open(settings.Format)
	if err != nil {
		return err
	}
	graph, err := f.Graph("Make driver")
	if err != nil {
		return err
	}
	env := map[string]any{
		"$load":        settings.LoadGlob,
		"$loadsep":     settings.LoadSeparator,
		"$save":        settings.SavePath,
		"$pkg":         settings.Pkg,
		"$prefix":      settings.Prefix,
		"$tableprefix": "",
		"$droptables":  false,
	}
	_, err = pipeline.RunExpr(graph, nil, env)
	return err
}

type MakeDriverSettings struct {
	// The desired storage format for the driver. Currently supported:
	// "sqlite"
	Format string

	// LoadGlob is a glob to a folder containing
	// domain classes that will be used to generate the driver.
	LoadGlob string

	// LoadSeparator is an optional separator string used
	// to split the LoadGlob if you want to provide multiple
	// globs.
	// Example:
	// LoadSeparator=";"
	// LoadGlob="/path/to/1;/path/to/2" will be evaulated
	// as two separate paths and the results combined.
	LoadSeparator string

	// SavePath is a filepath to a folder where the new
	// driver will be saved.
	SavePath string

	// Pkg is the name of the package to use for the new driver.
	Pkg string

	// Prefix is the prefix name to use for the driver types.
	Prefix string
}
