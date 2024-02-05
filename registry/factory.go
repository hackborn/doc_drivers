package registry

import (
	"fmt"
	"slices"

	"github.com/hackborn/doc_drivers/graphs"
)

type Factory struct {
	// Name is a unique key for this factory.
	Name string

	// DbPath is the location of the database for
	// any doc drivers.
	DbPath string

	// Open gets called when the factory is opened.
	Open OpenFunc

	graphEntries map[string]graphs.Entry
	graphNames   []string
}

func NewFactory(graphEntries map[string]graphs.Entry) Factory {
	graphNames := []string{}
	for k, _ := range graphEntries {
		graphNames = append(graphNames, k)
	}
	slices.Sort(graphNames)

	f := Factory{graphEntries: graphEntries, graphNames: graphNames}

	return f
}

func (f *Factory) GraphNames() []string {
	return f.graphNames
}

func (f *Factory) Graph(name string) (string, error) {
	if e, ok := f.graphEntries[name]; ok && e.Graph != nil {
		return e.Graph()
	}
	return "", fmt.Errorf("No graph for name \"%v\"", name)
}
