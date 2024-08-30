package nodes

import (
	"bytes"
	"cmp"
	"fmt"
	"go/format"
	"slices"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/hackborn/doc_drivers/enc"
	"github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
)

func newGoNode() pipeline.Node {
	caser := cases.Title(language.English)
	n := &goNode{caser: caser}
	return n
}

type goNode struct {
	goNodeSharedData

	caser cases.Caser
}

type goNodeSharedData struct {
	// Name of the package I'm writing into / generating.
	Pkg string

	// Prefix to use for my generated types.
	Prefix string

	// Optional prefix to prepend to table names. Only
	// used during driver development.
	TablePrefix string

	// Various configurable properties.
	Flags string
}

type goNodeData struct {
	goNodeSharedData

	casingFn casingFunc

	// Building -- this is the generated data that is
	// waiting to get flushed.
	templates map[string]string
	metadata  []MetadataDef
	json      []JsonDef
	// Store any structs that get new names, so we can
	// replace.
	jsonRenames map[string]string
}

func (n *goNode) Start(input pipeline.StartInput) error {
	data := goNodeData{casingFn: casingPassthrough}
	data.jsonRenames = make(map[string]string)
	data.goNodeSharedData = n.goNodeSharedData
	input.SetNodeData(&data)
	return nil
}

func (n *goNode) Run(state *pipeline.State, input pipeline.RunInput, output *pipeline.RunOutput) error {
	if err := n.initTemplates(state); err != nil {
		return err
	}
	eb := &errors.FirstBlock{}
	data := state.NodeData.(*goNodeData)
	if strings.Contains(strings.ToLower(data.Flags), "lowercase") {
		data.casingFn = casingLower
	}
	for _, pin := range input.Pins {
		switch p := pin.Payload.(type) {
		case *pipeline.StructData:
			eb.AddError(n.runStructPin(data, p))
		}
	}
	return eb.Err
}

func (n *goNode) Flush(state *pipeline.State, output *pipeline.RunOutput) error {
	data := state.NodeData.(*goNodeData)
	n.flushRenames(data)
	err := n.flushValidate(data)
	if err != nil {
		return err
	}
	vars, err := n.flushMakeVars(data)
	if err != nil {
		return fmt.Errorf("go node err: %w", err)
	}

	err = n.flushTemplates(data, vars, output)
	if err != nil {
		return fmt.Errorf("go node makeTemplates err: %w", err)
	}

	return err
}

func (n *goNode) initTemplates(state *pipeline.State) error {
	data := state.NodeData.(*goNodeData)
	if data.templates != nil {
		return nil
	}

	settings := templateSettings{
		fromPrefix: "_ref",
		toPrefix:   data.Prefix,
		pkg:        data.Pkg,
	}
	templates, err := makeTemplates(settings)
	if err != nil {
		return err
	}

	// Save for inspection
	/*
		for k, v := range templates {
			fn := "template_" + k + ".txt"
			err = os.WriteFile("template_"+k+".txt", []byte(v), 0600)
			fmt.Println("wrote template", fn, "err", err)
		}
	*/

	data.templates = templates
	return nil
}

func (n *goNode) runStructPin(data *goNodeData, pin *pipeline.StructData) error {
	md, jd, err := n.runMetadataDef(data, pin)
	if md.RootBucket != "-" {
		// If the skip indicator is false we add the metadata object.
		// Even if skip is true we will still add the json def, because
		// it might be used as a field inside an included metadate object.
		data.metadata = append(data.metadata, md)
	}
	data.json = append(data.json, jd)
	return err
}

func (n *goNode) runMetadataDef(data *goNodeData, pin *pipeline.StructData) (MetadataDef, JsonDef, error) {
	md := MetadataDef{DomainName: pin.Name}
	jd := JsonDef{Name: data.Prefix + "Json" + pin.Name}
	md.RootBucket = data.casingFn(pin.Name)
	md.NewConvStruct = jd.Name
	data.jsonRenames[pin.Name] = jd.Name

	for _, field := range pin.Fields {
		jf := JsonFieldDef{Name: field.Name, Type: field.RawType}
		// Default JSON tag. It may be replaced or cleared according
		// to the following rules.
		jsonTag := data.casingFn(field.Name)
		if field.Tag != "" {
			pt, err := enc.ParseTag(field.Tag)
			err = cmp.Or(err, pt.Validate())
			if err != nil {
				return md, jd, err
			}
			if pt.Name == "-" {
				// Omit this field from the DB.
				continue
			} else if pt.HasKey {
				if pt.Autoinc() && field.RawType != "uint64" {
					return md, jd, fmt.Errorf("Autoinc must be on uint64 type (%v/%v)", pin.Name, field.Name)
				}
				boltName := data.casingFn(field.Name)
				if pt.Name != "" {
					boltName = pt.Name
				}
				ft := "stringType"
				if field.RawType == "uint64" {
					ft = "uint64Type"
				}
				keyInfo := metadataKeyInfo{group: pt.KeyGroup, index: pt.KeyIndex}
				key := MetadataKeyDef{DomainName: field.Name,
					BoltName: boltName,
					Ft:       ft,
					Flags:    pt.Flags,
					keyInfo:  &keyInfo,
				}
				md.Buckets = append(md.Buckets, key)
				// Since this is a key it shouldn't be in the json
				jsonTag = ""
			} else {
				// Json tag has been assigned.
				if pt.Name != "" {
					jsonTag = pt.Name
				}
			}
		}
		// If there's no json tag, don't need a json field
		if jsonTag != "" {
			jf.Tag = "`json:" + `"` + jsonTag + `"` + "`"
			jd.Fields = append(jd.Fields, jf)
		}
	}

	for _, field := range pin.UnexportedFields {
		if field.Tag != "" {
			pt, err := enc.ParseTag(field.Tag)
			if err != nil {
				return md, jd, err
			}
			if pt.Name != "" {
				md.RootBucket = pt.Name
			}
		}
	}

	return md, jd, nil
}

func (n *goNode) flushRenames(data *goNodeData) {
	for _, jd := range data.json {
		for i, field := range jd.Fields {
			if r := renameType(field.Type, data.jsonRenames); r != "" {
				field.Type = r
				jd.Fields[i] = field
			}
		}
	}
}

func (n *goNode) flushValidate(nodeData *goNodeData) error {
	for _, m := range nodeData.metadata {
		err := m.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *goNode) flushMakeVars(nodeData *goNodeData) (map[string]any, error) {
	if nodeData.Pkg == "" {
		return nil, fmt.Errorf("Requires package name (set Pkg= on node)")
	}

	m := make(map[string]any)
	m["Prefix"] = nodeData.Prefix
	slices.SortFunc(nodeData.json, func(a, b JsonDef) int {
		return strings.Compare(a.Name, b.Name)
	})
	m["Json"] = nodeData.json
	slices.SortFunc(nodeData.metadata, func(a, b MetadataDef) int {
		return strings.Compare(a.DomainName, b.DomainName)
	})
	for i, md := range nodeData.metadata {
		slices.SortFunc(md.Buckets, func(a, b MetadataKeyDef) int {
			return compareKeys(a.keyInfo, b.keyInfo)
		})
		// Autoincs are always at the tail.
		md.sortAutoInc()
		(&md).setLeaf()
		nodeData.metadata[i] = md
	}
	m["Metadata"] = nodeData.metadata
	return m, nil
}

func (n *goNode) flushTemplates(nodeData *goNodeData, vars map[string]any, output *pipeline.RunOutput) error {
	eb := &errors.FirstBlock{}
	prefix := nodeData.Prefix
	if prefix != "" {
		prefix += "_"
	}
	for k, v := range nodeData.templates {
		b, err := n.runTemplate(v, vars)
		eb.AddError(err)
		b, err = n.runFormat(b)
		if err != nil {
			err = fmt.Errorf("FILE %v: %w", k, err)
		}
		eb.AddError(err)
		if err == nil {
			name := prefix + k + ".go"
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

// renameType will replace the type name if there is another one in the renames list.
// Empty string is returned if there is no replacement.
func renameType(typename string, renames map[string]string) string {
	if strings.HasPrefix(typename, "[]") {
		n := typename[2:]
		if r, ok := renames[n]; ok {
			return "[]" + r
		}
	} else if strings.HasPrefix(typename, "map[") {
		// TODO
	} else if r, ok := renames[typename]; ok {
		return r
	}
	return ""
}

type casingFunc func(string) string

func casingLower(s string) string {
	return strings.ToLower(s)
}

func casingPassthrough(s string) string {
	return s
}
