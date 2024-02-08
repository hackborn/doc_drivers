package nodes

import (
	"fmt"

	oferrors "github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
	ofstrings "github.com/hackborn/onefunc/strings"
)

func newSqlNode(tablePrefix string, dropTables bool) pipeline.Node {
	n := &sqlNode{Format: FormatSqlite, TablePrefix: tablePrefix, DropTables: dropTables}
	// Make functions
	n.makes = []makeSqlPinFunc{
		n.makeDefinitionPin,
	}
	return n
}

type sqlNode struct {
	Format      string
	TablePrefix string
	DropTables  bool
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
	eb := &oferrors.FirstBlock{}
	//	sb := ofstrings.GetWriter(eb)
	//	defer ofstrings.PutWriter(sb)

	md, err := makeMetadata(pin, n.TablePrefix)
	eb.AddError(err)

	cols := n.makeDefinitionCols(md, eb)
	create := n.makeDefinitionCreate(md, eb)
	def := cols + "\n" + create
	/*
		if n.DropTables {
			sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", md.Name))
		}
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
	*/

	content := &pipeline.ContentData{Name: pin.Name,
		Data:   def,
		Format: n.Format,
	}
	return pipeline.Pin{Name: definitionKey, Payload: content}, eb.Err
}

func (n *sqlNode) makeDefinitionCols(md metadata, eb oferrors.Block) string {
	sb := ofstrings.GetWriter(eb)
	defer ofstrings.PutWriter(sb)

	sb.WriteString("\tcols: []{{.Prefix}}SqlTableCol{\n")
	for _, field := range md.Fields {
		sqlType := convertGoTypeToSQLType(field.Type)
		sb.WriteString(fmt.Sprintf("\t\t{`%s`, `%s`},\n", field.Tag, sqlType))
	}

	sb.WriteString("\t},")

	/*
		cols: []_refSqlTableCol{
		{`id`, `VARCHAR(255)`},
							{`name`, `VARCHAR(255)`},
							{`val`, `INTEGER`},
							{`fy`, `INTEGER`},
						},
	*/

	return ofstrings.String(sb)
}

func (n *sqlNode) makeDefinitionCreate(md metadata, eb oferrors.Block) string {
	sb := ofstrings.GetWriter(eb)
	defer ofstrings.PutWriter(sb)

	sb.WriteString("create: `")
	if n.DropTables {
		sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", md.Name))
	}
	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", md.Name))

	for _, field := range md.Fields {
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

	sb.WriteString(");`,")

	return ofstrings.String(sb)
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
