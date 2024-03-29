package nodes

import (
	"bufio"
	"strings"

	"github.com/hackborn/doc_drivers/registry"
)

type MetadataDef struct {
	Name  string
	Value string
}

type TableDef struct {
	Name       string
	Statements string
}

type makeTemplateContent struct {
	registry.Content
	b        strings.Builder
	nodeData *templateNodeData
}

func newMakeTemplateContent(name, content string, data *templateNodeData) *makeTemplateContent {
	ic := &makeTemplateContent{nodeData: data}
	ic.Name = name
	ic.b.Grow(len(content) * 2)
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		ic.Lines = append(ic.Lines, scanner.Text())
	}
	return ic
}
