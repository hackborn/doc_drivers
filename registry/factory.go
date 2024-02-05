package registry

import (
	"fmt"
	"slices"

	"github.com/hackborn/doc_drivers/graphs"
)

type Factory struct {
	Name       string
	DriverName string
	DbPath     string

	Open OpenFunc

	// NewRef is a function to generate a new doc driver instance based on the reference driver.
	NewRef NewDriverFunc

	// NewGenerated generates a new doc driver based on the generated driver.
	NewGenerated NewDriverFunc

	// Clients can add additional processing when generating templates.
	ProcessTemplate ProcessTemplateFunc

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
