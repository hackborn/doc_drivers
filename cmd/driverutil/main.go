package main

import (
	"fmt"
	"path"
	"slices"
	"strings"

	oferrors "github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
	"github.com/manifoldco/promptui"

	_ "github.com/hackborn/doc_drivers"
	"github.com/hackborn/doc_drivers/registry"
)

func main() {
	f, err := getBackend()
	if err == quitErr {
		return
	}
	oferrors.LogFatal(err)

	graph, err := getGraph(f)
	if err == quitErr {
		return
	}
	oferrors.LogFatal(err)

	err = run(graph, f)
	oferrors.LogFatal(err)
}

func run(graph string, f registry.Factory) error {
	p, err := pipeline.Compile(graph)
	if err != nil {
		return err
	}
	// Supply the env vars
	env := makeEnv(p.Env(), f)
	_, err = pipeline.RunExpr(graph, nil, env)
	return err
}

func makeEnv(env map[string]any, f registry.Factory) map[string]any {
	if env == nil {
		env = make(map[string]any)
	}
	pathroot := path.Join("..", "..")
	for k, v := range env {
		if sv, ok := v.(string); ok {
			env[k] = strings.ReplaceAll(sv, "$pathroot", pathroot)
		}
	}
	env["$backend"] = f.Name
	// These are driver-development only settings, which should
	// be the only time this is getting hit.
	env["$tableprefix"] = "gen"
	env["$droptables"] = true
	return env
}

//local matches [../../domain/company.go ../../domain/events.go ../../domain/filing.go ../../domain/settings.go] <nil>

func getBackend() (registry.Factory, error) {
	backendNames := registry.Names()
	if len(backendNames) < 1 {
		return registry.Factory{}, fmt.Errorf("no backends available")
	}
	slices.Sort(backendNames)
	prompt := promptui.Select{
		Label: "Select a backend. Ctrl-C to quit",
		Items: backendNames,
	}
	_, backendName, err := prompt.Run()
	if err != nil {
		if err.Error() == `^C` {
			return registry.Factory{}, quitErr
		}
		return registry.Factory{}, fmt.Errorf("prompt error: %w", err)
	}

	return registry.Open(backendName)
}

func getGraph(f registry.Factory) (string, error) {
	graphNames := f.GraphNames()
	if len(graphNames) < 1 {
		return "", fmt.Errorf("no operations available")
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

	graph, err := f.Graph(graphName)
	if err != nil {
		return "", err
	}
	return graph, nil
}

var quitErr = fmt.Errorf("quit")
