package sqliterefdriver

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	oferrors "github.com/hackborn/onefunc/errors"
)

type _refSqlTableDef struct {
	cols []_refSqlTableCol
	// The SQL create string for this table.
	create string
}

func (d *_refSqlTableDef) Col(name string) (_refSqlTableCol, bool) {
	for _, col := range d.cols {
		if col.name == name {
			return col, true
		}
	}
	return _refSqlTableCol{}, false
}

type _refSqlTableCol struct {
	name   string
	dbType string
}

// _refRawSqlTable is a representation of an existing SQL table.
// It is an intermediary before building a local SQL table.
type _refRawSqlTable struct {
	Names []string
	Types []reflect.Type
}

func _refSqlSyncTable(db *sql.DB, name string, meta *_refMetadata, eb oferrors.Block) {
	constTable := _refTableDefs[name]
	// Always try and create it. This is important for testing, which wants
	// to construct the table each time
	_, err := db.Exec(constTable.create)
	eb.AddError(err)
	sqlTable := _refNewSqlTable(db, meta.table, eb)
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

func _refNewSqlTable(db *sql.DB, tablename string, eb oferrors.Block) _refSqlTableDef {
	// SQLite describe table
	stmt := `pragma table_info('` + tablename + `');`
	// What everyone else seems to use, need to verify format
	//	stmt := `DESCRIBE ` + tablename + `;`
	rows, err := db.Query(stmt)
	if err != nil {
		eb.AddError(err)
		return _refSqlTableDef{}
	}
	defer rows.Close()

	raw := _refNewRawSqlTable(rows, eb)
	if eb.HasError() {
		return _refSqlTableDef{}
	}

	table := _refNewSqlTableFromRaw(rows, raw, eb)
	return table
}

func _refNewRawSqlTable(rows *sql.Rows, eb oferrors.Block) _refRawSqlTable {
	types, err := rows.ColumnTypes()
	eb.AddError(err)
	cols, err := rows.Columns()
	eb.AddError(err)
	count := len(cols)
	if count != len(types) {
		eb.AddError(fmt.Errorf("SQL table columns and types length mismatch"))
	}
	if eb.HasError() {
		return _refRawSqlTable{}
	}

	resp := _refRawSqlTable{Names: cols}
	resp.Types = make([]reflect.Type, count, count)
	for i := 0; i < count; i++ {
		resp.Types[i] = types[i].ScanType()
	}
	return resp
}

func _refNewSqlTableFromRaw(rows *sql.Rows, raw _refRawSqlTable, eb oferrors.Block) _refSqlTableDef {
	count := len(raw.Names)
	dbName := ""
	dbType := ""
	var dest = make([]any, count, count)
	for i, _ := range dest {
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

	ans := _refSqlTableDef{}
	for rows.Next() {
		//if err = rows.Scan(&i64, &name, nil, nil, nil, nil); err != nil {
		if err := rows.Scan(dest...); err != nil {
			eb.AddError(err)
		}
		ans.cols = append(ans.cols, _refSqlTableCol{name: dbName, dbType: dbType})
	}
	return ans
}
