package sqlitebackend

import (
	"fmt"
	"strings"

	"github.com/hackborn/doc_drivers/registry"
	"github.com/hackborn/onefunc/errors"
)

func makeTemplates(c *registry.Content) error {
	eb := &errors.FirstBlock{}
	eb.AddError(makeTableDefs(c))
	return eb.Err
}

func makeTableDefs(c *registry.Content) error {
	if c.Name != "const" {
		return nil
	}
	s, e, ok := findRange(c, beginTableDefs, endTableDefs)
	if !ok {
		return fmt.Errorf("file \"" + c.Name + "\" is missing tabledefs")
	}
	lines := c.Lines[:s]
	lines = append(lines, tableDefsLine)
	lines = append(lines, c.Lines[e+1:]...)
	c.Lines = lines
	return nil
}

func findRange(c *registry.Content, start, end string) (int, int, bool) {
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
	tableDefsLine  = "{{range .Tabledefs}}`{{.Name}}`: `{{.Statements}}`,{{end}}"
)
