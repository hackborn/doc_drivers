package main

import (
	"fmt"
	"path/filepath"

	"github.com/hackborn/doc_drivers"
	_ "github.com/hackborn/doc_drivers/cmd/makeDriver/nodes"
	"github.com/hackborn/doc_drivers/registry"
	"github.com/hackborn/onefunc/pipeline"
	_ "github.com/hackborn/onefunc/pipeline/nodes"
	"github.com/hackborn/tox"
	"github.com/hackborn/tox/togo"
)

// main generates the reference templates.
func main() {
	f, err := drivers.GetFactoryFromCla()
	if err != nil {
		fmt.Println(err)
		fmt.Println(help)
		return
	}
	//	err = runToxGenerate(f)
	err = runGenerate(f)
	if err != nil {
		fmt.Println(err)
		fmt.Println(help)
		return
	}
}

func runGenerate(f registry.Factory) error {
	loadPath := filepath.Join("..", "..", "domain", "*")
	savePath := filepath.Join("data")
	expr := `graph (load(Glob="` + loadPath + `") -> struct -> go(Pkg=db, Prefix=test) -> fmt -> save(Path="` + savePath + `"))`
	_, err := pipeline.RunExpr(expr, nil)
	return err
}

func runToxGenerate(f registry.Factory) error {
	src := tox.SrcFolder(filepath.Join("..", "..", "domain"))
	gonode := togo.NewNode(togo.WithFormat("sqlite"), togo.WithPackageName("db"), togo.WithPrefix("_tst"))
	foldernode := tox.NewFolderNode(filepath.Join("data"))
	return tox.Run(src,
		//		tosql.NewNode(),
		tox.NewFmtNode(),
		gonode,
		// tox.NewFmtNode(),
		foldernode)
}

// {{.package}}, {{.toxPackage}}, {{.prefix}}
// tableDefsLine  = "{{range .tabledefs}}`{{.name}}`: `{{.statements}}`,{{end}}"
