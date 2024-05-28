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
				if outpin.Payload != nil {
					output.Pins = append(output.Pins, outpin)
				}
			}
		}
	}
	return nil
}

func (n *sqlNode) makeDefinitionPin(data *sqlNodeData, state *pipeline.State, pin *pipeline.StructData) (pipeline.Pin, error) {
	md, ok, err := makeMetadata(pin, data.TablePrefix)
	if !ok {
		return pipeline.Pin{}, nil
	}
	eb := &oferrors.FirstBlock{}
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
	keys := md.KeySpecs()
	for _, field := range md.Fields {
		format := ""
		// Masks are defined in ref/ref_sql.go
		mask := "0"
		sqlType := convertGoTypeToSQLType(field.Type)
		if field.Type == pipeline.UnknownType {
			format = "json"
		}
		// In SQL, integer primary keys are always auto increment.
		if sqlType == sqlInteger && keys.isPrimary(field.Tag) {
			mask = "colFlagAuto"
		}
		sb.WriteString(fmt.Sprintf("\t\t{`%s`, `%s`, `%s`, %s},\n", field.Tag, sqlType, format, mask))
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

	keys := md.KeySpecs()

	for i, field := range md.Fields {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sqlType := convertGoTypeToSQLType(field.Type)
		postFix := ""
		if keys.isPrimary(field.Tag) {
			postFix += " NOT NULL"
		}
		sb.WriteString(fmt.Sprintf("\t%s %s%s", field.Tag, sqlType, postFix))
	}

	// The first key is the primary, the rest are indexes
	for i, groupSpec := range keys.keyGroups {
		columnNames := groupSpec.columnNames()
		ca := ofstrings.CompileArgs{Separator: ","}
		s := ofstrings.CompileStrings(ca, columnNames...)

		if i == 0 {
			sb.WriteString(",\n")
			sb.WriteString("\tPRIMARY KEY (" + s + ")\n);\n")
		} else if groupSpec.name != "" {
			sb.WriteString(fmt.Sprintf("CREATE INDEX IF NOT EXISTS %v ON %v (%v);\n", groupSpec.name, md.Name, s))
		}
	}

	sb.WriteString("`,")

	return ofstrings.String(sb)
}

// convertGoTypeToSQLType converts a Go type to an SQL data type.
func convertGoTypeToSQLType(goType string) string {
	switch goType {
	case "string":
		return "VARCHAR(255)"
	case "int", "int64":
		return sqlInteger
	case "float", "float64":
		return "FLOAT"
	case "bool":
		return "BOOLEAN"
	default:
		return "TEXT"
	}
}

const sqlInteger = "INTEGER"

const (
	// NOTE: Flags are replicated in ref/ref_sql.go
	colFlagAuto = 1 << iota // The column value is auto-generated.
)
