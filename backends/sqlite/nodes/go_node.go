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
	n := &goNode{caser: caser}
	// Currently sqlite is the only supported format, so I'll make it a default
	n.Format = FormatSqlite
	return n
}

type goNode struct {
	goNodeSharedData

	caser cases.Caser
}

type goNodeSharedData struct {
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
}

type goNodeData struct {
	goNodeSharedData

	// Building -- this is the generated data that is
	// waiting to get flushed.
	structs     map[string]*pipeline.StructData
	definitions map[string]string
	metadata    map[string]string
}

func (n *goNodeData) fileName(base, format string) string {
	prefix := strings.Trim(n.Prefix, "_")
	if prefix != "" {
		prefix += "_"
	}

	// Trim the extension
	base = base[:len(base)-len(path.Ext(base))]

	return prefix + base + format
}

func (n *goNode) Start(input pipeline.StartInput) error {
	data := goNodeData{}
	data.goNodeSharedData = n.goNodeSharedData
	data.structs = make(map[string]*pipeline.StructData)
	data.definitions = make(map[string]string)
	data.metadata = make(map[string]string)
	input.SetNodeData(&data)
	return nil
}

func (n *goNode) Run(state *pipeline.State, input pipeline.RunInput, output *pipeline.RunOutput) error {
	data := state.NodeData.(*goNodeData)
	eb := &errors.FirstBlock{}
	for _, pin := range input.Pins {
		switch p := pin.Payload.(type) {
		case *pipeline.StructData:
			eb.AddError(n.runStructPin(data, state, p))
		}
	}
	return eb.Err
}

func (n *goNode) Flush(state *pipeline.State, output *pipeline.RunOutput) error {
	data := state.NodeData.(*goNodeData)
	vars, err := n.makeVars(data)
	if err != nil {
		return fmt.Errorf("go node err: %w", err)
	}

	//	fmt.Println("vars", vars)
	err = n.makeTemplates(data, vars, output)
	if err != nil {
		return fmt.Errorf("go node makeTemplates err: %w", err)
	}
	return err
}

func (n *goNode) runStructPin(data *goNodeData, state *pipeline.State, pin *pipeline.StructData) error {
	data.structs[pin.Name] = pin
	switch data.Format {
	case FormatSqlite:
		return n.runStructPinSqlite(data, state, pin)
	default:
		return fmt.Errorf("go node: Unknown format \"%v\"", data.Format)
	}
}

func (n *goNode) runStructPinSqlite(nodeData *goNodeData, state *pipeline.State, pin *pipeline.StructData) error {
	/*
		sn := newSqlNode(nodeData.TablePrefix, nodeData.DropTables)
		output := &pipeline.RunOutput{}
		err := sn.Run(state, pipeline.NewRunInput(pipeline.Pin{Payload: pin}), output)
		if err != nil {
			return err
		}
	*/
	sn := newSqlNode(nodeData.TablePrefix, nodeData.DropTables)
	output := &pipeline.RunOutput{}
	err := pipeline.RunNode(sn, pipeline.NewRunInput(pipeline.Pin{Payload: pin}), output)
	if err != nil {
		return err
	}

	// Metadata
	eb := errors.FirstBlock{}
	nodeData.metadata[pin.Name] = n.makeMetadataValue(nodeData, pin, &eb)
	if eb.Err != nil {
		return eb.Err
	}

	// Table definitions
	for _, pin := range output.Pins {
		switch p := pin.Payload.(type) {
		case *pipeline.ContentData:
			if pin.Name == definitionKey {
				if _, ok := nodeData.definitions[p.Name]; ok {
					return fmt.Errorf("go node supplied multiple definitions with the same name (%v)", p.Name)
				}
				data := strings.ReplaceAll(p.Data, "{{.Prefix}}", nodeData.Prefix)
				nodeData.definitions[p.Name] = data
			}
		}
	}
	return nil
}

func (n *goNode) makeMetadataValue(nodeData *goNodeData, pin *pipeline.StructData, eb errors.Block) string {
	w := ofstrings.GetWriter(eb)
	defer ofstrings.PutWriter(w)
	ca := ofstrings.CompileArgs{Quote: "\"", Separator: ",", Eb: eb}
	md, err := makeMetadata(pin, nodeData.TablePrefix)
	eb.AddError(err)
	fn := md.FieldNames()
	tn := md.TagNames()

	w.WriteString("&" + nodeData.Prefix + "Metadata{\n")
	w.WriteString("\t\t\ttable: \"" + md.Name + "\",\n")
	w.WriteString("\t\t\ttags: []string{" + ofstrings.CompileStrings(ca, tn...) + "},\n")
	w.WriteString("\t\t\tfields: []string{" + ofstrings.CompileStrings(ca, fn...) + "},\n")
	w.WriteString("\t\t\tkeys: map[string]*" + nodeData.Prefix + "KeyMetadata{\n")
	for k, _ := range md.Keys {
		w.WriteString("\t\t\t\t\"" + k + "\": &" + nodeData.Prefix + "KeyMetadata{\n")
		w.WriteString("\t\t\t\t\ttags: []string{" + ofstrings.CompileStrings(ca, md.KeyTagNames(k)...) + "},\n")
		w.WriteString("\t\t\t\t\tfields: []string{" + ofstrings.CompileStrings(ca, md.KeyFieldNames(k)...) + "},\n")
		w.WriteString("\t\t\t\t},\n")
	}
	w.WriteString("\t\t\t},\n")
	return ofstrings.String(w)
}

func (n *goNode) makeTemplates(nodeData *goNodeData, vars map[string]any, output *pipeline.RunOutput) error {
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
			name := nodeData.fileName(path.Base(match), ".go")
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

func (n *goNode) makeVars(nodeData *goNodeData) (map[string]any, error) {
	if nodeData.Format == "" {
		return nil, fmt.Errorf("Requires format (set Format= on node)")
	}
	if nodeData.Pkg == "" {
		return nil, fmt.Errorf("Requires package name (set Pkg= on node)")
	}

	m := make(map[string]any)
	m["Package"] = nodeData.Pkg
	m["UtilPackage"] = registry.UtilPackageName
	m["Prefix"] = nodeData.Prefix
	var tableDefs []TableDef
	for k, v := range nodeData.definitions {
		tableDefs = append(tableDefs, TableDef{Name: k, Statements: v})
	}
	m["Tabledefs"] = tableDefs
	var metadatas []MetadataDef
	for k, v := range nodeData.metadata {
		metadatas = append(metadatas, MetadataDef{Name: k, Value: v})
	}
	m["Metadata"] = metadatas
	m["Datestamp"] = time.Now().Format(time.DateOnly)
	return m, nil
}
