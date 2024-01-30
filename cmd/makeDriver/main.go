package main

import (
	"fmt"
	"path/filepath"

	"github.com/hackborn/doc_drivers"
	_ "github.com/hackborn/doc_drivers/cmd/makeDriver/nodes"
	"github.com/hackborn/doc_drivers/registry"
	"github.com/hackborn/onefunc/pipeline"
	_ "github.com/hackborn/onefunc/pipeline/nodes"
)

// main generates the reference templates.
func main() {
	f, err := drivers.GetFactoryFromCla()
	if err != nil {
		fmt.Println(err)
		fmt.Println(help)
		return
	}
	err = runGenerate(f)
	if err != nil {
		fmt.Println(err)
		fmt.Println(help)
		return
	}
}

func runGenerate(f registry.Factory) error {
	env := map[string]any{`$load`: filepath.Join("..", "..", "domain", "*"),
		`$save`: filepath.Join("..", "..", "drivers", "sqlite", "gen"),
		//		`$save`:   filepath.Join("data"),
		`$pkg`:    "sqlitegendriver",
		`$prefix`: "gen"}
	expr := `graph (load(Glob=$load) -> struct(Tag=tox) -> go(Pkg=$pkg, Prefix=$prefix) -> save(Path=$save))`
	_, err := pipeline.RunExpr(expr, nil, env)
	return err
}
