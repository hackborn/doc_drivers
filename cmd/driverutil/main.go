package main

import (
	"fmt"

	"github.com/hackborn/doc_drivers"
	_ "github.com/hackborn/doc_drivers/nodes"
	"github.com/hackborn/doc_drivers/registry"
	oferrors "github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
	_ "github.com/hackborn/onefunc/pipeline/nodes"
	"github.com/manifoldco/promptui"
)

func main() {
	graph, err := getGraph()
	if err == quitErr {
		return
	}
	oferrors.LogFatal(err)

	f, err := getDriver()
	if err == quitErr {
		return
	}
	oferrors.LogFatal(err)

	err = run(graph, f)
	oferrors.LogFatal(err)
}

func run(graph string, f registry.Factory) error {
	env := map[string]any{
		"$drivername": f.Name,
	}
	_, err := pipeline.RunExpr(graph, nil, env)
	return err
}

func getGraph() (string, error) {
	graphNames := drivers.GraphNames()
	if len(graphNames) < 1 {
		return "", fmt.Errorf("no graphs available")
	}
	prompt := promptui.Select{
		Label: "Select an operation. Ctrl-C to quit",
		Items: graphNames,
	}
	_, graphName, err := prompt.Run()
	if err != nil {
		if err.Error() == `^C` {
			return "", quitErr
		}
		return "", fmt.Errorf("prompt error: %w", err)
	}

	graph, err := drivers.Graph(graphName)
	if err != nil {
		return "", err
	}
	return graph, nil
}

func getDriver() (registry.Factory, error) {
	driverNames := drivers.DriverNames()
	if len(driverNames) < 1 {
		return registry.Factory{}, fmt.Errorf("no drivers available")
	}
	prompt := promptui.Select{
		Label: "Select a driver. Ctrl-C to quit",
		Items: driverNames,
	}
	_, driverName, err := prompt.Run()
	if err != nil {
		if err.Error() == `^C` {
			return registry.Factory{}, quitErr
		}
		return registry.Factory{}, fmt.Errorf("prompt error: %w", err)
	}

	if f, ok := registry.Find(driverName); ok {
		return f, nil
	}
	return registry.Factory{}, fmt.Errorf("No driver available for name \"%v\"", driverName)
}

var quitErr = fmt.Errorf("quit")
