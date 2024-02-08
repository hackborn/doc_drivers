package sqlitegendriver

// autogenerated with github.com/hackborn/doc_drivers on 2024-02-08
// do not modify

const (
	genKeysVar        = "$KEYS$"
	genKeyValuesVar   = "$KEYVALUES$"
	genFieldsVar      = "$FIELDS$"
	genTableVar       = "$TABLE$"
	genValuesVar      = "$VALUES$"
	genFieldValuesVar = "$FIELDVALUES$"

	genAndKeyword      = "AND"
	genAndKeywordWS    = " AND "
	genEqualsKeyword   = "="
	genEqualsKeywordWS = " = "

	genQuoteSz = string(rune('\''))

	genSetSql = `INSERT INTO $TABLE$ ($FIELDS$) VALUES($VALUES$) ON CONFLICT($KEYS$) DO UPDATE SET $FIELDVALUES$;`
	genDelSql = `DELETE FROM $TABLE$ WHERE ($KEYVALUES$);`
)

var (
	genTableDefs = map[string]genSqlTableDef{
		`Company`: {
			cols: []genSqlTableCol{
				{`id`, `VARCHAR(255)`},
				{`name`, `VARCHAR(255)`},
				{`val`, `INTEGER`},
				{`fy`, `INTEGER`},
			},
			create: `DROP TABLE IF EXISTS gencompany;
CREATE TABLE IF NOT EXISTS gencompany (
	id VARCHAR(255),
	name VARCHAR(255),
	val INTEGER,
	fy INTEGER,
	PRIMARY KEY (id)
);`,
		}, `Filing`: {
			cols: []genSqlTableCol{
				{`ticker`, `VARCHAR(255)`},
				{`end`, `VARCHAR(255)`},
				{`form`, `VARCHAR(255)`},
				{`val`, `INTEGER`},
				{`units`, `VARCHAR(255)`},
				{`fy`, `INTEGER`},
			},
			create: `DROP TABLE IF EXISTS genfiling;
CREATE TABLE IF NOT EXISTS genfiling (
	ticker VARCHAR(255),
	end VARCHAR(255),
	form VARCHAR(255),
	val INTEGER,
	units VARCHAR(255),
	fy INTEGER,
	PRIMARY KEY (ticker,end,form)
);`,
		},
	}

	genMetadatas = map[string]*genMetadata{
		`Company`: &genMetadata{
			table:  "gencompany",
			tags:   []string{"id", "name", "val", "fy"},
			fields: []string{"Id", "Name", "Value", "FoundedYear"},
			keys: map[string]*genKeyMetadata{
				"b": &genKeyMetadata{
					tags:   []string{"name"},
					fields: []string{"Name"},
				},
				"c": &genKeyMetadata{
					tags:   []string{"fy"},
					fields: []string{"FoundedYear"},
				},
				"": &genKeyMetadata{
					tags:   []string{"id"},
					fields: []string{"Id"},
				},
			},
		}, `Filing`: &genMetadata{
			table:  "genfiling",
			tags:   []string{"ticker", "end", "form", "val", "units", "fy"},
			fields: []string{"Ticker", "EndDate", "Form", "Value", "Units", "FiscalYear"},
			keys: map[string]*genKeyMetadata{
				"": &genKeyMetadata{
					tags:   []string{"ticker", "end", "form"},
					fields: []string{"Ticker", "EndDate", "Form"},
				},
			},
		},
	}
)
