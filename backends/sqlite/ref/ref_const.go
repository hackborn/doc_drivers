package sqliterefdriver

const (
	_refKeysVar        = "$KEYS$"
	_refKeyValuesVar   = "$KEYVALUES$"
	_refFieldsVar      = "$FIELDS$"
	_refTableVar       = "$TABLE$"
	_refValuesVar      = "$VALUES$"
	_refFieldValuesVar = "$FIELDVALUES$"

	_refAndKeyword      = "AND"
	_refAndKeywordWS    = " AND "
	_refEqualsKeyword   = "="
	_refEqualsKeywordWS = " = "

	_refQuoteSz = string(rune('\''))

	_refSetSql = `INSERT INTO $TABLE$ ($FIELDS$) VALUES($VALUES$) ON CONFLICT($KEYS$) DO UPDATE SET $FIELDVALUES$;`
	_refDelSql = `DELETE FROM $TABLE$ WHERE ($KEYVALUES$);`
)

var (
	_refTableDefs = map[string]_refSqlTableDef{
		// Begin tabledefs
		`Company`: {
			cols: []_refSqlTableCol{
				{`id`, `VARCHAR(255)`},
				{`name`, `VARCHAR(255)`},
				{`val`, `INTEGER`},
				{`fy`, `INTEGER`},
			},
			create: `DROP TABLE IF EXISTS refcompany;
CREATE TABLE IF NOT EXISTS refcompany (
		id VARCHAR(255),
		name VARCHAR(255),
		val INTEGER,
		fy INTEGER,
		PRIMARY KEY (id)
);`,
		},
		`Filing`: {
			cols: []_refSqlTableCol{
				{`ticker`, `VARCHAR(255)`},
				{`end`, `VARCHAR(255)`},
				{`form`, `VARCHAR(255)`},
				{`val`, `INTEGER`},
				{`units`, `VARCHAR(255)`},
				{`fy`, `INTEGER`},
			},
			create: `DROP TABLE IF EXISTS reffiling;
CREATE TABLE IF NOT EXISTS reffiling (
		ticker VARCHAR(255),
		end VARCHAR(255),
		form VARCHAR(255),
		val INTEGER,
		units VARCHAR(255),
		fy INTEGER,
		PRIMARY KEY (ticker, end, form)
);`,
		},
		// End tabledefs
	}

	_refMetadatas = map[string]*_refMetadata{
		// Begin metadata
		"Company": &_refMetadata{
			table:  "refcompany",
			tags:   []string{"id", "name", "val", "fy"},
			fields: []string{"Id", "Name", "Value", "FoundedYear"},
			keys: map[string]*_refKeyMetadata{
				"": &_refKeyMetadata{
					tags:   []string{"id"},
					fields: []string{"Id"},
				},
				"b": &_refKeyMetadata{
					tags:   []string{"name"},
					fields: []string{"Name"},
				},
			},
		},
		"Filing": &_refMetadata{
			table:  "reffiling",
			tags:   []string{"ticker", "end", "form", "val", "units", "fy"},
			fields: []string{"Ticker", "EndDate", "Form", "Value", "Units", "FiscalYear"},
			keys: map[string]*_refKeyMetadata{
				"": &_refKeyMetadata{
					tags:   []string{"ticker", "end", "form"},
					fields: []string{"Ticker", "EndDate", "Form"},
				},
			},
		},
		// End metadata
	}
)
