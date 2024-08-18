package nodes

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hackborn/doc_drivers/registry"
	"github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
)

func newTemplateNode() pipeline.Node {
	n := &templateNode{}
	// Process functions
	n.processors = []makeTemplateContentFunc{
		n.processPrefix,
		n.processPackage,
		n.processAutogenerated,
		n.processTableDefs,
		n.processMetadata,
	}
	return n
}

type templateNode struct {
	templateNodeSharedData

	processors []makeTemplateContentFunc
}

type templateNodeSharedData struct {
	Prefix string
}

type templateNodeData struct {
	templateNodeSharedData

	// A collection of keys made by the processors.
	// This isn't used for anything; used to be to
	// report what keys were created, but that's
	// not being done anymore.
	mapped map[string]string
}

func (n *templateNode) Start(input pipeline.StartInput) error {
	data := templateNodeData{}
	data.templateNodeSharedData = n.templateNodeSharedData
	data.mapped = make(map[string]string)
	input.SetNodeData(&data)
	return nil
}

func (n *templateNode) Run(state *pipeline.State, input pipeline.RunInput, output *pipeline.RunOutput) error {
	data := state.NodeData.(*templateNodeData)
	eb := &errors.FirstBlock{}
	for _, pin := range input.Pins {
		switch p := pin.Payload.(type) {
		case *pipeline.ContentData:
			eb.AddError(n.runContentPin(data, p, output))
		}
	}
	return eb.Err
}

func (n *templateNode) runContentPin(data *templateNodeData, pin *pipeline.ContentData, output *pipeline.RunOutput) error {
	// XXX ideally everything is hooked up enough to check the format is "go"
	if pin.Name == "" {
		return fmt.Errorf("template node: missing pin name")
	}
	ic := newMakeTemplateContent(pin.Name, pin.Data, data)
	for _, fn := range n.processors {
		err := fn(ic)
		if err != nil {
			return err
		}
	}
	ic.b.Reset()
	for _, line := range ic.Lines {
		ic.b.WriteString(line)
		ic.b.WriteString("\n")
	}
	outpin := &pipeline.ContentData{Name: pin.Name + ".txt", Data: ic.b.String(), Format: "txt"}
	output.Pins = append(output.Pins, pipeline.Pin{Payload: outpin})
	return nil
}

func (n *templateNode) processPrefix(c *makeTemplateContent) error {
	if c.nodeData.Prefix == "" {
		return nil
	}
	for i, line := range c.Lines {
		c.Lines[i] = strings.ReplaceAll(line, c.nodeData.Prefix, templatePrefixKey)
	}
	c.nodeData.mapped[templatePrefixKey] = ""
	return nil
}

func (n *templateNode) processPackage(c *makeTemplateContent) error {
	const prefix string = "package "
	for i, line := range c.Lines {
		if strings.HasPrefix(line, prefix) {
			c.Lines[i] = prefix + templatePackageKey
			c.nodeData.mapped[templatePackageKey] = ""
			return nil
		}
	}
	return fmt.Errorf("no package line")
}

func (n *templateNode) processAutogenerated(c *makeTemplateContent) error {
	const prefix string = `// autogenerated with `
	const line0 string = ""
	const line1 string = prefix + templateUtilPackageKey
	const line2 string = `// do not modify`
	for i, line := range c.Lines {
		if strings.HasPrefix(line, prefix) {
			c.Lines[i] = line1
			return nil
		}
	}

	if len(c.Lines) < 2 {
		c.Lines = append(c.Lines, []string{line0, line1, line2}...)
	} else {
		c.Lines = slices.Insert(c.Lines, 1, line0, line1, line2)
	}
	c.nodeData.mapped[templateUtilPackageKey] = ""
	c.nodeData.mapped[templateDatestampKey] = ""
	return nil
}

func (n *templateNode) processTableDefs(c *makeTemplateContent) error {
	if c.Name != "const" {
		return nil
	}
	s, e, ok := findContentRange(&c.Content, beginTableDefs, endTableDefs)
	if !ok {
		return fmt.Errorf("file \"" + c.Name + "\" is missing tabledefs comment")
	}
	lines := c.Lines[:s]
	lines = append(lines, tableDefsLine)
	lines = append(lines, c.Lines[e+1:]...)
	c.Lines = lines
	return nil
}

func (n *templateNode) processMetadata(c *makeTemplateContent) error {
	if c.Name != "const" {
		return nil
	}
	s, e, ok := findContentRange(&c.Content, beginMetadata, endMetadata)
	if !ok {
		return fmt.Errorf("file \"" + c.Name + "\" is missing metadata comment")
	}
	lines := c.Lines[:s]
	lines = append(lines, metadataLine)
	lines = append(lines, c.Lines[e+1:]...)
	c.Lines = lines
	return nil
}

func findContentRange(c *registry.Content, start, end string) (int, int, bool) {
	sp := -1
	ep := -1
	for i, s := range c.Lines {
		if strings.Index(s, start) >= 0 {
			sp = i
		}
		if strings.Index(s, end) >= 0 {
			ep = i
		}
		if sp >= 0 && ep >= 0 {
			return sp, ep, true
		}
	}
	return -1, -1, false
}

const (
	beginTableDefs = `// Begin tabledefs`
	endTableDefs   = `// End tabledefs`
	tableDefsLine  = "{{range .Tabledefs}}`{{.Name}}`: {\n{{.Statements}}\n},{{end}}"

	beginMetadata = `// Begin metadata`
	endMetadata   = `// End metadata`
	metadataLine  = "{{range .Metadata}}`{{.Name}}`: {{.Value}}},{{end}}"
)
