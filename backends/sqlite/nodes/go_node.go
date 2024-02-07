package nodes

import (
	"bytes"
	"fmt"
	"go/format"
	"io/fs"
	"path"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/hackborn/doc_drivers/registry"
	"github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
	ofstrings "github.com/hackborn/onefunc/strings"
)

func newGoNode() pipeline.Node {
	caser := cases.Title(language.English)
	// Currently sqlite is the only supported format, so I'll make it a default
	return &goNode{Format: FormatSqlite,
		caser: caser}
}

type goNode struct {
	// Format of the driver I'm generating. Currently only sqlite is
	// supported, so that's used as the default.
	Format string

	// Name of the package I'm writing into / generating.
	Pkg string

	// Prefix to use for my generated types.
	Prefix string

	// Optional prefix to prepend to table names. Only
	// used during driver development.
	TablePrefix string

	// If true, the existing tables will be dropped and
	// created anew with each driver run. Only used during
	// development.
	DropTables bool

	caser cases.Caser
}

type goNodeState struct {
	// Building -- this is the generated data that is
	// waiting to get flushed.
	structs     map[string]*pipeline.StructData
	definitions map[string]string
	metadata    map[string]string
}

func getGoNodeState(key pipeline.Node, state *pipeline.State) *goNodeState {
	ns := pipeline.GetNodeState[goNodeState](key, state)
	if ns.structs == nil {
		ns.structs = make(map[string]*pipeline.StructData)
	}
	if ns.definitions == nil {
		ns.definitions = make(map[string]string)
	}
	if ns.metadata == nil {
		ns.metadata = make(map[string]string)
	}
	return ns
}

func (n *goNode) Run(state *pipeline.State, input pipeline.RunInput) (*pipeline.RunOutput, error) {
	nodeState := getGoNodeState(n, state)
	eb := &errors.FirstBlock{}
	for _, pin := range input.Pins {
		switch p := pin.Payload.(type) {
		case *pipeline.StructData:
			eb.AddError(n.runStructPin(nodeState, state, p))
		}
	}
	return nil, eb.Err
}

func (n *goNode) Flush(state *pipeline.State) (*pipeline.RunOutput, error) {
	nodeState := getGoNodeState(n, state)
	vars, err := n.makeVars(nodeState)
	if err != nil {
		return nil, fmt.Errorf("go node err: %w", err)
	}

	output := &pipeline.RunOutput{}
	//	fmt.Println("vars", vars)
	err = n.makeTemplates(vars, output)
	return output, err
}

func (n *goNode) runStructPin(nodeState *goNodeState, state *pipeline.State, pin *pipeline.StructData) error {
	nodeState.structs[pin.Name] = pin
	switch n.Format {
	case FormatSqlite:
		return n.runStructPinSqlite(nodeState, state, pin)
	default:
		return fmt.Errorf("go node: Unknown format \"%v\"", n.Format)
	}
}

func (n *goNode) runStructPinSqlite(nodeState *goNodeState, state *pipeline.State, pin *pipeline.StructData) error {
	sn := newSqlNode(n.TablePrefix, n.DropTables)
	output, err := sn.Run(state, pipeline.NewInput(pipeline.Pin{Payload: pin}))
	if err != nil {
		return err
	}

	// Metadata
	eb := errors.FirstBlock{}
	nodeState.metadata[pin.Name] = n.makeMetadataValue(pin, &eb)
	if eb.Err != nil {
		return eb.Err
	}

	// Table definitions
	for _, pin := range output.Pins {
		switch p := pin.Payload.(type) {
		case *pipeline.ContentData:
			if pin.Name == definitionKey {
				if _, ok := nodeState.definitions[p.Name]; ok {
					return fmt.Errorf("togo supplied multiple definitions with the same name (%v)", p.Name)
				}
				nodeState.definitions[p.Name] = p.Data
			}
		}
	}
	return nil
}

func (n *goNode) makeMetadataValue(pin *pipeline.StructData, eb errors.Block) string {
	w := ofstrings.GetWriter(eb)
	defer ofstrings.PutWriter(w)
	ca := ofstrings.CompileArgs{Quote: "\"", Separator: ",", Eb: eb}
	md, err := makeMetadata(pin, n.TablePrefix)
	eb.AddError(err)
	fn := md.FieldNames()
	tn := md.TagNames()

	w.WriteString("&" + n.Prefix + "Metadata{\n")
	w.WriteString("\t\t\ttable: \"" + md.Name + "\",\n")
	w.WriteString("\t\t\ttags: []string{" + ofstrings.CompileStrings(ca, tn...) + "},\n")
	w.WriteString("\t\t\tfields: []string{" + ofstrings.CompileStrings(ca, fn...) + "},\n")
	w.WriteString("\t\t\tkeys: map[string]*" + n.Prefix + "KeyMetadata{\n")
	for k, _ := range md.Keys {
		w.WriteString("\t\t\t\t\"" + k + "\": &" + n.Prefix + "KeyMetadata{\n")
		w.WriteString("\t\t\t\t\ttags: []string{" + ofstrings.CompileStrings(ca, md.KeyTagNames(k)...) + "},\n")
		w.WriteString("\t\t\t\t\tfields: []string{" + ofstrings.CompileStrings(ca, md.KeyFieldNames(k)...) + "},\n")
		w.WriteString("\t\t\t\t},\n")
	}
	w.WriteString("\t\t\t},\n")
	return ofstrings.String(w)
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

func (n *goNode) makeVars(nodeState *goNodeState) (map[string]any, error) {
	if n.Format == "" {
		return nil, fmt.Errorf("Requires format (set Format= on node)")
	}
	if n.Pkg == "" {
		return nil, fmt.Errorf("Requires package name (set Pkg= on node)")
	}

	m := make(map[string]any)
	m["Package"] = n.Pkg
	m["UtilPackage"] = registry.UtilPackageName
	m["Prefix"] = n.Prefix
	var tableDefs []TableDef
	for k, v := range nodeState.definitions {
		tableDefs = append(tableDefs, TableDef{Name: k, Statements: v})
	}
	m["Tabledefs"] = tableDefs
	var metadatas []MetadataDef
	for k, v := range nodeState.metadata {
		metadatas = append(metadatas, MetadataDef{Name: k, Value: v})
	}
	m["Metadata"] = metadatas
	m["Datestamp"] = time.Now().Format(time.DateOnly)
	return m, nil
}
