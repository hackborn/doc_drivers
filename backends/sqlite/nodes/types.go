package nodes

import (
	"bufio"
	"strings"

	"github.com/hackborn/doc_drivers/registry"
)

type TableDef struct {
	Name       string
	Statements string
}

type makeTemplateContent struct {
	registry.Content
	b strings.Builder
}

func newMakeTemplateContent(name, content string) *makeTemplateContent {
	ic := &makeTemplateContent{}
	ic.Name = name
	ic.b.Grow(len(content) * 2)
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		ic.Lines = append(ic.Lines, scanner.Text())
	}
	return ic
}
