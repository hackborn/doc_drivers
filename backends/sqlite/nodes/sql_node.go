package nodes

import (
	"fmt"

	oferrors "github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
	ofstrings "github.com/hackborn/onefunc/strings"
)

func newSqlNode(tablePrefix string) pipeline.Node {
	n := &sqlNode{Format: FormatSqlite, TablePrefix: tablePrefix}
	// Make functions
	n.makes = []makeSqlPinFunc{
		n.makeDefinitionPin,
	}
	return n
}

type sqlNode struct {
	Format      string
	TablePrefix string
	makes       []makeSqlPinFunc
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

	md, err := makeMetadata(pin, n.TablePrefix)
	block.AddError(err)

	// Debugging
	sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", md.Name))
	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", md.Name))

	// Iterate over struct fields
	for _, field := range md.Fields {
		// Convert Go type to SQL data type
		sqlType := convertGoTypeToSQLType(field.Type)
		sb.WriteString(fmt.Sprintf("\t%s %s,\n", field.Tag, sqlType))
	}

	// XXX Figure out how sqlite does non-primary keys
	if pk, ok := md.PrimaryKey(); ok {
		keyNames := md.KeyTagNames(pk)
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
