package sqlitegendriver

// autogenerated with github.com/hackborn/onefunc/pipeline on 01 Feb 24 21:20 PST
// do not modify

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/hackborn/doc"
	"github.com/hackborn/onefunc/assign"
	"github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/extract"
	ofstrings "github.com/hackborn/onefunc/strings"
)

type genDriver struct {
	db            *sql.DB
	sqlDriverName string
	format        doc.Format
}

func (d *genDriver) Open(dataSourceName string) (doc.Driver, error) {
	eb := &errors.FirstBlock{}
	db, err := sql.Open(d.sqlDriverName, dataSourceName)
	eb.AddError(err)
	eb.AddError(d.runDefinitions(db))
	if err != nil {
		return nil, err
	}
	f := doc.FormatWithDefaults(genNewFormat())
	return &genDriver{db: db, format: f}, nil
}

func (d *genDriver) Close() error {
	db := d.db
	d.db = nil
	if db != nil {
		return db.Close()
	}
	return nil
}

func (d *genDriver) Format() doc.Format {
	return d.format
}

func (d *genDriver) Set(req doc.SetRequestAny, a doc.Allocator) (*doc.Optional, error) {
	eb := &errors.FirstBlock{}
	meta, ok := genMetadatas[a.TypeName()]
	if !ok {
		return nil, fmt.Errorf("missing metadata for \"%v\"", a.TypeName())
	}
	keys, ok := meta.keys[""]
	if !ok {
		return nil, fmt.Errorf("missing primary key metadata for \"%v\"", a.TypeName())
	}

	statement := genSetSql
	handler := &fieldsAndValuesHandler{}
	ca1 := ofstrings.CompileArgs{Quote: "", Separator: ", ", Eb: eb}
	ca2 := ofstrings.CompileArgs{Quote: genQuoteSz, Separator: ", ", Eb: eb}
	extract.From(req.ItemAny(), extract.NewChain(meta.FieldsToTags(), handler))
	s := strings.ReplaceAll(statement, genFieldsVar, ofstrings.Compile(ca1, handler.fields...))
	s = strings.ReplaceAll(s, genValuesVar, ofstrings.Compile(ca2, handler.values...))
	s = strings.ReplaceAll(s, genFieldValuesVar, makeExcludedFieldValues(eb, handler.fields))
	s = strings.ReplaceAll(s, genTableVar, meta.table)
	s = strings.ReplaceAll(s, genKeysVar, ofstrings.CompileStrings(ca1, keys.tags...))
	if eb.Err != nil {
		return nil, eb.Err
	}

	//	fmt.Println("EXEC", s)
	if _, err := d.db.Exec(s); err != nil {
		return nil, err
	}
	return nil, nil
}

func (d *genDriver) Get(req doc.GetRequest, a doc.Allocator) (*doc.Optional, error) {
	eb := &errors.FirstBlock{}
	meta, ok := genMetadatas[a.TypeName()]
	if !ok {
		return nil, fmt.Errorf("missing metadata for \"%v\"", a.TypeName())
	}
	tags, fields := getTagsAndFields(meta, req)
	if len(tags) < 1 {
		return nil, fmt.Errorf("missing fields for \"%v\"", a.TypeName())
	}
	ca := ofstrings.CompileArgs{Quote: "", Separator: ", ", Eb: eb}
	selectFields := ofstrings.CompileStrings(ca, tags...)
	where, err := whereClause(req)
	if eb.Err != nil {
		return nil, eb.Err
	}
	if err != nil {
		return nil, err
	}
	s := "SELECT " + selectFields + " FROM " + meta.table + where + ";"
	//	fmt.Println("QUERY 1", s)
	rows, err := d.db.Query(s)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fieldCount := len(tags)
	var dest = make([]any, fieldCount, fieldCount)
	for i, _ := range dest {
		dest[i] = new(any)
	}

	vreq := assign.ValuesRequest{
		FieldNames: fields,
		NewValues:  dest,
	}

	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			return nil, err
		}
		resp := a.New()
		if err = assign.Values(vreq, resp); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (d *genDriver) Delete(req doc.DeleteRequestAny, a doc.Allocator) (*doc.Optional, error) {
	meta, ok := genMetadatas[a.TypeName()]
	if !ok {
		return nil, fmt.Errorf("missing metadata for \"%v\"", a.TypeName())
	}
	keys, ok := meta.keys[""]
	if !ok {
		return nil, fmt.Errorf("missing primary key metadata for \"%v\"", a.TypeName())
	}

	opts := extract.SliceOpts{Assign: doc.AssignKeyword, Combine: doc.AndKeyword}
	exprSlice := extract.AsSlice(req.ItemAny(), extract.NewChain(keys.FieldsToTags()), &opts)
	expr := ""
	dexpr, err := doc.NewExpr(d.format, exprSlice...)
	if err != nil {
		return nil, err
	} else {
		s, err := dexpr.Format()
		if err != nil {
			return nil, err
		}
		expr = s
	}

	s := genDelSql
	s = strings.ReplaceAll(s, genTableVar, meta.table)
	s = strings.ReplaceAll(s, genKeyValuesVar, expr)
	// fmt.Println("delete statemet", s)

	if _, err := d.db.Exec(s); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *genDriver) runDefinitions(db *sql.DB) error {
	for _, v := range genDefinitions {
		if _, err := db.Exec(v); err != nil {
			return err
		}
	}
	return nil
}