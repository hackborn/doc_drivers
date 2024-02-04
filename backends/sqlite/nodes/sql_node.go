package nodes

import (
	"fmt"
	"slices"
	"strings"

	oferrors "github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
	ofstrings "github.com/hackborn/onefunc/strings"
)

func newSqlNode() pipeline.Runner {
	n := &sqlNode{Format: formatSqlite}
	// Make functions
	n.makes = []makeSqlPinFunc{
		n.makeDefinitionPin,
	}
	return n
}

type sqlNode struct {
	Format string
	makes  []makeSqlPinFunc
}

func (n *sqlNode) Run(state *pipeline.State, input pipeline.RunInput) (*pipeline.RunOutput, error) {
	output := &pipeline.RunOutput{}
	for _, pin := range input.Pins {
		switch p := pin.Payload.(type) {
		case *pipeline.StructData:
			for _, fn := range n.makes {
				outpin, err := fn(state, p)
				if err != nil {
					return nil, err
				}
				output.Pins = append(output.Pins, outpin)
			}
		}
	}
	return output, nil
}

func (n *sqlNode) makeDefinitionPin(state *pipeline.State, pin *pipeline.StructData) (pipeline.Pin, error) {
	block := oferrors.FirstBlock{}
	sb := ofstrings.GetWriter(&block)
	defer ofstrings.PutWriter(sb)

	// Debugging
	sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", pin.Name))
	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", pin.Name))

	// Iterate over struct fields
	for _, field := range pin.Fields {
		// The name should be the tag name, but default to field name
		pt := newSqlParsedTag(field)
		fieldName := pt.name
		if fieldName == "" {
			fieldName = field.Name
		}

		// Convert Go type to SQL data type
		fieldType := field.Type
		sqlType := convertGoTypeToSQLType(fieldType)

		sb.WriteString(fmt.Sprintf("\t%s %s,\n", fieldName, sqlType))
	}

	keyNames := sqlKeyNames(pin)
	if len(keyNames) > 0 {
		ca := ofstrings.CompileArgs{Separator: ","}
		keys := ofstrings.CompileStrings(ca, keyNames...)
		sb.WriteString("\tPRIMARY KEY (" + keys + ")\n")
	}

	sb.WriteString(");")
	content := &pipeline.ContentData{Name: pin.Name,
		Data:   ofstrings.String(sb),
		Format: n.Format,
	}
	return pipeline.Pin{Name: definitionKey, Payload: content}, block.Err
}

// convertGoTypeToSQLType converts a Go type to an SQL data type.
func convertGoTypeToSQLType(goType string) string {
	switch goType {
	case "string":
		return "VARCHAR(255)"
	case "int", "int64":
		return "INTEGER"
	case "float", "float64":
		return "FLOAT"
	case "bool":
		return "BOOLEAN"
	default:
		return "TEXT"
	}
}

// sqlKeys answers the fields that have all been tagged with a "key",
// in sorted order. The name answered will either be the tag name, if present,
// or the field name.
func sqlKeyNames(sd *pipeline.StructData) []string {
	if len(sd.Fields) < 1 {
		return []string{}
	}
	keys := make([]sqlKey, 0, len(sd.Fields))
	for _, field := range sd.Fields {
		pt := newSqlParsedTag(field)
		if pt.key != "" {
			keys = append(keys, sqlKey{tagName: pt.name, fieldName: field.Name, keyName: pt.key})
		}
	}
	if len(keys) < 1 {
		return []string{}
	}
	slices.SortFunc(keys, func(a, b sqlKey) int {
		return strings.Compare(a.keyName, b.keyName)
	})
	var keyNames []string
	for _, k := range keys {
		if k.tagName != "" {
			keyNames = append(keyNames, k.tagName)
		} else {
			keyNames = append(keyNames, k.fieldName)
		}
	}
	return keyNames
}

type sqlKey struct {
	tagName   string
	fieldName string
	keyName   string
}

type sqlParsedTag struct {
	name string
	key  string
}

// newSqlParsedTag takes a field tag and extracts the name
// and key state.
func newSqlParsedTag(f pipeline.StructField) sqlParsedTag {
	pt := sqlParsedTag{}
	tags := strings.Split(f.Tag, ",")
	for i, t := range tags {
		t = strings.TrimSpace(t)
		if i == 0 {
			pt.name = t
		} else if strings.HasPrefix(t, "key") {
			pt.key = t
		}
	}
	return pt
}
