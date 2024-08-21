package registry

import (
	"github.com/hackborn/doc"
)

// OpenFunc is called on a factory when a backend is loaded.
type OpenFunc func() error

// NewDriverFunc returns a new instance of a doc.Driver.
type NewDriverFunc func() doc.Driver

// PrepareRunFunc provides factories a change to make any
// changes before running a graph.
type PrepareRunFunc func(f Factory, graphName string, vars map[string]any)

type ProcessTemplateFunc func(*Content) error
