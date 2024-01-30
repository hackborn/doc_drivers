package drivers

import (
	"fmt"
	"os"

	_ "github.com/hackborn/doc_drivers/drivers/sqlite"
	"github.com/hackborn/doc_drivers/registry"
)

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
