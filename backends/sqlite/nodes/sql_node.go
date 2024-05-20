package nodes

import (
	"fmt"

	oferrors "github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
	ofstrings "github.com/hackborn/onefunc/strings"
)

func newSqlNode(tablePrefix string, dropTables bool) pipeline.Node {
	n := &sqlNode{}
	n.sqlNodeData = sqlNodeData{Format: FormatSqlite, TablePrefix: tablePrefix, DropTables: dropTables}
	// Make functions
	n.makes = []makeSqlPinFunc{
		n.makeDefinitionPin,
	}
	return n
}

type sqlNode struct {
	sqlNodeData

	makes []makeSqlPinFunc
}

type sqlNodeData struct {
	Format      string
	TablePrefix string
	DropTables  bool
}

func (n *sqlNode) Start(input pipeline.StartInput) error {
	data := n.sqlNodeData
	input.SetNodeData(&data)
	return nil
}

func (n *sqlNode) Run(state *pipeline.State, input pipeline.RunInput, output *pipeline.RunOutput) error {
	data := state.NodeData.(*sqlNodeData)
	for _, pin := range input.Pins {
		switch p := pin.Payload.(type) {
		case *pipeline.StructData:
			for _, fn := range n.makes {
				outpin, err := fn(data, state, p)
				if err != nil {
					return err
				}
				output.Pins = append(output.Pins, outpin)
			}
		}
	}
	return nil
}

func (n *sqlNode) makeDefinitionPin(data *sqlNodeData, state *pipeline.State, pin *pipeline.StructData) (pipeline.Pin, error) {
	eb := &oferrors.FirstBlock{}
	//	sb := ofstrings.GetWriter(eb)
	//	defer ofstrings.PutWriter(sb)

	md, err := makeMetadata(pin, data.TablePrefix)
	eb.AddError(err)

	cols := n.makeDefinitionCols(md, eb)
	create := n.makeDefinitionCreate(md, eb)
	def := cols + "\n" + create

	content := &pipeline.ContentData{Name: pin.Name,
		Data:   def,
		Format: data.Format,
	}
	return pipeline.Pin{Name: definitionKey, Payload: content}, eb.Err
}

func (n *sqlNode) makeDefinitionCols(md metadata, eb oferrors.Block) string {
	sb := ofstrings.GetWriter(eb)
	defer ofstrings.PutWriter(sb)

	sb.WriteString("\tcols: []{{.Prefix}}SqlTableCol{\n")
	for _, field := range md.Fields {
		sqlType := convertGoTypeToSQLType(field.Type)
		format := ""
		if field.Type == pipeline.UnknownType {
			format = "json"
		}
		sb.WriteString(fmt.Sprintf("\t\t{`%s`, `%s`, `%s`},\n", field.Tag, sqlType, format))
	}

	sb.WriteString("\t},")

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
