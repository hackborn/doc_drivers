package sqliterefdriver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/hackborn/doc"
	"github.com/hackborn/onefunc/assign"
	"github.com/hackborn/onefunc/errors"
	oferrors "github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/extract"
	ofstrings "github.com/hackborn/onefunc/strings"
)

type _refDriver struct {
	db            *sql.DB
	sqlDriverName string
	format        doc.Format
}

func (d *_refDriver) Open(dataSourceName string) (doc.Driver, error) {
	eb := &errors.FirstBlock{}
	db, err := sql.Open(d.sqlDriverName, dataSourceName)
	eb.AddError(err)
	eb.AddError(d.syncTables(db))
	if eb.Err != nil {
		return nil, eb.Err
	}
	f := doc.FormatWithDefaults(_refNewFormat())
	return &_refDriver{db: db, format: f}, nil
}

func (d *_refDriver) Close() error {
	db := d.db
	d.db = nil
	if db != nil {
		return db.Close()
	}
	return nil
}

func (d *_refDriver) Format() doc.Format {
	return d.format
}

func (d *_refDriver) Set(req doc.SetRequestAny, a doc.Allocator) (*doc.Optional, error) {
	meta, keys, cols, err := d.prepareSet(a)
	if err != nil {
		return nil, err
	}

	eb := &errors.FirstBlock{}
	statement := _refSetSql
	handler := &fieldsAndValuesHandler{cols: cols, filter: req.GetFilter()}
	ca1 := ofstrings.CompileArgs{Quote: "", Separator: ", ", Eb: eb}
	ca2 := ofstrings.CompileArgs{Quote: _refQuoteSz, Separator: ", ", Eb: eb}
	extract.From(req.ItemAny(), extract.NewChain(meta.FieldsToTags(), handler))
	s := strings.ReplaceAll(statement, _refFieldsVar, ofstrings.Compile(ca1, handler.fields...))
	s = strings.ReplaceAll(s, _refValuesVar, ofstrings.Compile(ca2, handler.values...))
	s = strings.ReplaceAll(s, _refFieldValuesVar, makeExcludedFieldValues(eb, handler.fields))
	s = strings.ReplaceAll(s, _refTableVar, meta.table)
	s = strings.ReplaceAll(s, _refKeysVar, ofstrings.CompileStrings(ca1, keys.tags...))
	if eb.Err != nil {
		return nil, eb.Err
	}

	//	fmt.Println("EXEC", s)
	if _, err := d.db.Exec(s); err != nil {
		return nil, err
	}
	return nil, nil
}

func (d *_refDriver) prepareSet(a doc.Allocator) (*_refMetadata, *_refKeyMetadata, []_refSqlTableCol, error) {
	tn := a.TypeName()
	meta, ok := _refMetadatas[tn]
	if !ok {
		return nil, nil, nil, fmt.Errorf("missing metadata for \"%v\"", tn)
	}
	keys, ok := meta.keys[""]
	if !ok {
		return nil, nil, nil, fmt.Errorf("missing primary key metadata for \"%v\"", tn)
	}
	tableDef, ok := _refTableDefs[tn]
	if !ok {
		return nil, nil, nil, fmt.Errorf("missing tabledef for \"%v\"", tn)
	}
	return meta, keys, tableDef.cols, nil
}

func (d *_refDriver) Get(req doc.GetRequest, a doc.Allocator) (*doc.Optional, error) {
	meta, tags, fields, assigns, err := d.prepareGet(req, a)
	if err != nil {
		return nil, err
	}
	eb := &errors.FirstBlock{}
	ca := ofstrings.CompileArgs{Quote: "", Separator: ", ", Eb: eb}
	selectFields := ofstrings.CompileStrings(ca, tags...)
	where, err := whereClause(req)
	if eb.Err != nil {
		return nil, eb.Err
	}
	if err != nil {
		return nil, err
	}
	s := "SELECT "
	if req.Flags&doc.GetUnique != 0 {
		s += "DISTINCT "
	}
	s += selectFields + " FROM " + meta.table + where + ";"
	// fmt.Println("QUERY 1", s)
	rows, err := d.db.Query(s)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fieldCount := len(tags)
	var dest = make([]any, fieldCount, fieldCount)
	for i := range dest {
		dest[i] = new(any)
	}

	vreq := assign.ValuesRequest{
		FieldNames: fields,
		NewValues:  dest,
		Assigns:    assigns,
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

func (d *_refDriver) prepareGet(req doc.GetRequest, a doc.Allocator) (*_refMetadata, []string, []string, []assign.AssignFunc, error) {
	tn := a.TypeName()
	meta, ok := _refMetadatas[tn]
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("missing metadata for \"%v\"", tn)
	}
	tags, fields := getTagsAndFields(meta, req)
	if len(tags) < 1 {
		return nil, nil, nil, nil, fmt.Errorf("missing fields for \"%v\"", tn)
	}
	tableDef, ok := _refTableDefs[tn]
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("missing tabledef for \"%v\"", tn)
	}
	assigns := tableDef.AssignsFor(tags)
	return meta, tags, fields, assigns, nil
}

func (d *_refDriver) Delete(req doc.DeleteRequestAny, a doc.Allocator) (*doc.Optional, error) {
	meta, ok := _refMetadatas[a.TypeName()]
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

	s := _refDelSql
	s = strings.ReplaceAll(s, _refTableVar, meta.table)
	s = strings.ReplaceAll(s, _refKeyValuesVar, expr)
	// fmt.Println("delete statemet", s)

	if _, err := d.db.Exec(s); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *_refDriver) syncTables(db *sql.DB) error {
	eb := &oferrors.FirstBlock{}
	for k, v := range _refMetadatas {
		_refSqlSyncTable(db, k, v, eb)
	}
	return eb.Err
}
