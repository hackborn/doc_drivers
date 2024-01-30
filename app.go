package drivers

import (
	"fmt"
	"os"

	_ "github.com/hackborn/doc_drivers/drivers/sqlite"
	"github.com/hackborn/doc_drivers/registry"
)

func DriverNames() []string {
	return registry.DriverNames()
}

func GraphNames() []string {
	return graphNames
}

func Graph(name string) (string, error) {
	entry, ok := graphEntries[name]
	if !ok {
		return "", fmt.Errorf("no entry for graph \"%v\"", name)
	}
	dat, err := graphs.ReadFile(entry)
	if err != nil {
		return "", fmt.Errorf("error: %w", err)
	}
	return string(dat), err
}

// GetFactoryFromCla answers the factory specified from the command line args.
func GetFactoryFromCla() (registry.Factory, error) {
	n := getDriverName()
	if n == "" {
		return registry.Factory{}, fmt.Errorf("No driver specified. Available drivers are %v", registry.DriverNames())
	}
	if f, ok := registry.Find(n); ok {
		return f, nil
	}
	return registry.Factory{}, fmt.Errorf("No driver specified for name %v. Available drivers are %v", n, registry.DriverNames())
}

func getDriverName() string {
	if len(os.Args) < 2 {
		return ""
	}
	return os.Args[1]
}

var (
	graphNames []string
	// A map of the friendly graph name to the entry name in the FS.
	graphEntries = make(map[string]string)
)
