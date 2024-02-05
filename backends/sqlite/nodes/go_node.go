package nodes

import (
	"bytes"
	"fmt"
	"go/format"
	"io/fs"
	"path"
	"reflect"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
)

func newGoNode() pipeline.Node {
	caser := cases.Title(language.English)
	structs := make(map[string]*pipeline.StructData)
	d := make(map[string]string)
	// Currently sqlite is the only supported format, so I'll make it a default
	return &goNode{Format: FormatSqlite,
		caser:       caser,
		definitions: d,
		structs:     structs}
}

type goNode struct {
	// Format of the driver I'm generating. Currently only sqlite is
	// supported, so that's used as the default.
	Format string

	// Name of the package I'm writing into / generating.
	Pkg string

	// Prefix to use for my generated types.
	Prefix string

	caser cases.Caser

	// Building -- this is the generated data that is
	// waiting to get flushed.
	structs     map[string]*pipeline.StructData
	definitions map[string]string
}

func (n *goNode) Run(state *pipeline.State, input pipeline.RunInput) (*pipeline.RunOutput, error) {
	eb := &errors.FirstBlock{}
	for _, pin := range input.Pins {
		switch p := pin.Payload.(type) {
		case *pipeline.StructData:
			eb.AddError(n.runStructPin(state, p))
		}
	}
	return nil, eb.Err
}

func (n *goNode) Flush(state *pipeline.State) (*pipeline.RunOutput, error) {
	vars, err := n.makeVars()
	if err != nil {
		return nil, fmt.Errorf("go node err: %w", err)
	}

	output := &pipeline.RunOutput{}
	//	fmt.Println("vars", vars)
	err = n.makeTemplates(vars, output)
	return output, err
}

func (n *goNode) runStructPin(state *pipeline.State, pin *pipeline.StructData) error {
	n.structs[pin.Name] = pin
	switch n.Format {
	case FormatSqlite:
		return n.runStructPinSqlite(state, pin)
	default:
		return fmt.Errorf("go node: Unknown format \"%v\"", n.Format)
	}
}

func (n *goNode) runStructPinSqlite(state *pipeline.State, pin *pipeline.StructData) error {
	sn := newSqlNode()
	output, err := sn.Run(state, pipeline.NewInput(pipeline.Pin{Payload: pin}))
	if err != nil {
		return err
	}
	for _, pin := range output.Pins {
		switch p := pin.Payload.(type) {
		case *pipeline.ContentData:
			if pin.Name == definitionKey {
				if _, ok := n.definitions[p.Name]; ok {
					return fmt.Errorf("togo supplied multiple definitions with the same name (%v)", p.Name)
				}
				n.definitions[p.Name] = p.Data
			}
		}
	}
	return nil
}

func (n *goNode) makeTemplates(vars map[string]any, output *pipeline.RunOutput) error {
	templatesFs, ok := pipeline.FindFs(TemplateFsName)
	if !ok {
		return fmt.Errorf("go node: No FS for name \"%v\"", TemplateFsName)
	}

	eb := &errors.FirstBlock{}
	matches, err := fs.Glob(templatesFs, "templates/*.txt")
	eb.AddError(err)

	for _, match := range matches {
		d, err := fs.ReadFile(templatesFs, match)
		eb.AddError(err)
		b, err := n.runTemplate(string(d), vars)
		eb.AddError(err)
		b, err = n.runFormat(b)
		eb.AddError(err)
		if err == nil {
			name := n.fileName(path.Base(match), ".go")
			output.Pins = append(output.Pins, pipeline.Pin{Payload: &pipeline.ContentData{Name: name, Data: string(b)}})
		}
	}
	return eb.Err
}

func (n *goNode) runTemplate(content string, vars map[string]any) ([]byte, error) {
	eb := errors.FirstBlock{}
	t1 := template.New("t1")
	t1, err := t1.Parse(content)
	eb.AddError(err)
	var buf bytes.Buffer
	err = t1.Execute(&buf, vars)
	eb.AddError(err)
	return buf.Bytes(), eb.Err
}

func (n *goNode) runFormat(src []byte) ([]byte, error) {
	return format.Source(src)
}

func (n *goNode) fileName(base, format string) string {
	prefix := strings.Trim(n.Prefix, "_")
	if prefix != "" {
		prefix += "_"
	}

	// Trim the extension
	base = base[:len(base)-len(path.Ext(base))]

	return prefix + base + format
}

func (n *goNode) makeVars() (map[string]any, error) {
	if n.Format == "" {
		return nil, fmt.Errorf("Requires format (set Format= on node)")
	}
	if n.Pkg == "" {
		return nil, fmt.Errorf("Requires package name (set Pkg= on node)")
	}

	// {{.package}}, {{.toxPackage}}, {{.prefix}}
	// {{range .tabledefs}}`{{.name}}`: `{{.statements}}`,{{end}}

	m := make(map[string]any)
	m["Package"] = n.Pkg
	m["ToxPackage"] = reflect.TypeOf(pipeline.ContentData{}).PkgPath()
	m["Prefix"] = n.Prefix
	var tableDefs []TableDef
	for k, v := range n.definitions {
		tableDefs = append(tableDefs, TableDef{Name: k, Statements: v})
	}
	m["Tabledefs"] = tableDefs
	m["Datestamp"] = time.Now().Format(time.RFC822)
	return m, nil
}
