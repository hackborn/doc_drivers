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
		`UiSetting`: {
			cols: []_refSqlTableCol{
				{`name`, `VARCHAR(255)`, ``},
				{`value`, `TEXT`, `json`},
			},
			create: `DROP TABLE IF EXISTS gensettings;
CREATE TABLE IF NOT EXISTS gensettings (
	name VARCHAR(255),
	value TEXT,
	PRIMARY KEY (name)
);`,
		}, `Company`: {
			cols: []_refSqlTableCol{
				{`id`, `VARCHAR(255)`, ``},
				{`name`, `VARCHAR(255)`, ``},
				{`val`, `INTEGER`, ``},
				{`fy`, `INTEGER`, ``},
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
			cols: []_refSqlTableCol{
				{`ticker`, `VARCHAR(255)`, ``},
				{`end`, `VARCHAR(255)`, ``},
				{`form`, `VARCHAR(255)`, ``},
				{`val`, `INTEGER`, ``},
				{`units`, `VARCHAR(255)`, ``},
				{`fy`, `INTEGER`, ``},
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
		}, `CollectionSetting`: {
			cols: []_refSqlTableCol{
				{`name`, `VARCHAR(255)`, ``},
				{`value`, `TEXT`, `json`},
			},
			create: `DROP TABLE IF EXISTS gensettings;
CREATE TABLE IF NOT EXISTS gensettings (
	name VARCHAR(255),
	value TEXT,
	PRIMARY KEY (name)
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
		`CollectionSetting`: &_refMetadata{
			table:  "genCollectionSetting",
			tags:   []string{"name", "value"},
			fields: []string{"Name", "Value"},
			keys: map[string]*_refKeyMetadata{
				"": &_refKeyMetadata{
					tags:   []string{"name"},
					fields: []string{"Name"},
				},
			},
		},
		`UiSetting`: &_refMetadata{
			table:  "genCollectionSetting",
			tags:   []string{"name", "value"},
			fields: []string{"Name", "Value"},
			keys: map[string]*_refKeyMetadata{
				"": &_refKeyMetadata{
					tags:   []string{"name"},
					fields: []string{"Name"},
				},
			},
		},
		// End metadata
	}
)
