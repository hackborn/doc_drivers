package sqlitegendriver

// autogenerated with github.com/hackborn/doc_drivers on 2024-06-02
// do not modify

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	oferrors "github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/values"
)

type genSqlTableDef struct {
	cols []genSqlTableCol
	// The SQL create string for this table.
	create string
}

func (d *genSqlTableDef) Col(name string) (genSqlTableCol, bool) {
	for _, col := range d.cols {
		if col.name == name {
			return col, true
		}
	}
	return genSqlTableCol{}, false
}

func (d *genSqlTableDef) AssignsFor(tags []string) []values.SetFunc {
	ans := make([]values.SetFunc, 0, len(tags))
	for _, tag := range tags {
		ans = append(ans, d.assignForTag(tag, d.cols))
	}
	return ans
}

func (d *genSqlTableDef) assignForTag(tag string, cols []genSqlTableCol) values.SetFunc {
	switch d.formatForTag(tag, cols) {
	case "json":
		return values.SetJson
	default:
		return nil
	}
}

func (d *genSqlTableDef) formatForTag(tag string, cols []genSqlTableCol) string {
	for _, col := range cols {
		if col.name == tag {
			return col.format
		}
	}
	return ""
}

type genSqlTableCol struct {
	// The column name.
	name string

	// The data type in the database.
	dbType string

	// format optionally specifies a serialization format for storing this
	// field in the database. For example, if the Go type can't be translated
	// to a type in the database, this can be specified to "json" to write
	// the value to JSON and store it as a string.
	format string

	// Additional info about this column.
	flags uint64
}

const (
	// NOTE: Flag names are referenced in nodes/sql_node.go
	colFlagAuto = 1 << iota // The column value is auto-generated.
)

// genRawSqlTable is a representation of an existing SQL table.
// It is an intermediary before building a local SQL table.
type genRawSqlTable struct {
	Names []string
	Types []reflect.Type
}

func genSqlSyncTable(db *sql.DB, name string, meta *genMetadata, eb oferrors.Block) {
	constTable := genTableDefs[name]
	// Always try and create it. This is important for testing, which wants
	// to construct the table each time
	_, err := db.Exec(constTable.create)
	eb.AddError(err)
	sqlTable := genNewSqlTable(db, meta.table, eb)
	if eb.HasError() {
		return
	}
	// We now have the field name and types, both what's the sql table and
	// what I have defined. Handle differences. We won't delete fields,
	// only add missing ones or error on changed ones.
	for _, constcol := range constTable.cols {
		if sqlcol, ok := sqlTable.Col(constcol.name); !ok {
			// Add the field
			stmt := `ALTER TABLE ` + meta.table + ` ADD COLUMN ` + constcol.name + ` ` + constcol.dbType + `;`
			_, err := db.Exec(stmt)
			eb.AddError(err)
		} else if constcol.dbType != sqlcol.dbType {
			eb.AddError(fmt.Errorf("You must manually edit the database. Column \"%v.%v\" was \"%v\" but needs to be \"%v\"", meta.table, constcol.name, sqlcol.dbType, constcol.dbType))
		}
	}
}

func genNewSqlTable(db *sql.DB, tablename string, eb oferrors.Block) genSqlTableDef {
	// SQLite describe table
	stmt := `pragma table_info('` + tablename + `');`
	// What everyone else seems to use, need to verify format
	//	stmt := `DESCRIBE ` + tablename + `;`
	rows, err := db.Query(stmt)
	if err != nil {
		eb.AddError(err)
		return genSqlTableDef{}
	}
	defer rows.Close()

	raw := genNewRawSqlTable(rows, eb)
	if eb.HasError() {
		return genSqlTableDef{}
	}

	table := genNewSqlTableFromRaw(rows, raw, eb)
	return table
}

func genNewRawSqlTable(rows *sql.Rows, eb oferrors.Block) genRawSqlTable {
	types, err := rows.ColumnTypes()
	eb.AddError(err)
	cols, err := rows.Columns()
	eb.AddError(err)
	count := len(cols)
	if count != len(types) {
		eb.AddError(fmt.Errorf("SQL table columns and types length mismatch"))
	}
	if eb.HasError() {
		return genRawSqlTable{}
	}

	resp := genRawSqlTable{Names: cols}
	resp.Types = make([]reflect.Type, count, count)
	for i := 0; i < count; i++ {
		resp.Types[i] = types[i].ScanType()
	}
	return resp
}

func genNewSqlTableFromRaw(rows *sql.Rows, raw genRawSqlTable, eb oferrors.Block) genSqlTableDef {
	count := len(raw.Names)
	dbName := ""
	dbType := ""
	var dest = make([]any, count, count)
	for i := range dest {
		lc := strings.ToLower(raw.Names[i])
		// Always provide a default, in case we hit a path that doesn't set a value.
		dest[i] = new(any)
		// Not sure how much variation there is between SQL
		// implementations but I've seen at least these.
		if lc == "name" || lc == "field" {
			if raw.Types[i].Kind() != reflect.String {
				eb.AddError(fmt.Errorf("name column is not string"))
			}
			dest[i] = &dbName
		} else if lc == "type" {
			if raw.Types[i].Kind() != reflect.String {
				eb.AddError(fmt.Errorf("type column is not string"))
			}
			dest[i] = &dbType
		}
	}

	ans := genSqlTableDef{}
	for rows.Next() {
		//if err = rows.Scan(&i64, &name, nil, nil, nil, nil); err != nil {
		if err := rows.Scan(dest...); err != nil {
			eb.AddError(err)
		}
		ans.cols = append(ans.cols, genSqlTableCol{name: dbName, dbType: dbType})
	}
	return ans
}
