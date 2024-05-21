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
				{`id`, `VARCHAR(255)`, ``, 0},
				{`name`, `VARCHAR(255)`, ``, 0},
				{`val`, `INTEGER`, ``, 0},
				{`fy`, `INTEGER`, ``, 0},
			},
			create: `DROP TABLE IF EXISTS gencompany;
CREATE TABLE IF NOT EXISTS gencompany (
	id VARCHAR(255) NOT NULL,
	name VARCHAR(255),
	val INTEGER,
	fy INTEGER,
	PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS b ON gencompany (name);
CREATE INDEX IF NOT EXISTS c ON gencompany (fy);
`,
		}, `Events`: {
			cols: []_refSqlTableCol{
				{`time`, `INTEGER`, ``, colFlagAuto},
				{`name`, `VARCHAR(255)`, ``, 0},
				{`value`, `VARCHAR(255)`, ``, 0},
			},
			create: `DROP TABLE IF EXISTS genevents;
CREATE TABLE IF NOT EXISTS genevents (
	time INTEGER NOT NULL,
	name VARCHAR(255),
	value VARCHAR(255),
	PRIMARY KEY (time)
);
`,
		}, `Filing`: {
			cols: []_refSqlTableCol{
				{`ticker`, `VARCHAR(255)`, ``, 0},
				{`end`, `VARCHAR(255)`, ``, 0},
				{`form`, `VARCHAR(255)`, ``, 0},
				{`val`, `INTEGER`, ``, 0},
				{`units`, `VARCHAR(255)`, ``, 0},
				{`fy`, `INTEGER`, ``, 0},
			},
			create: `DROP TABLE IF EXISTS genfiling;
CREATE TABLE IF NOT EXISTS genfiling (
	ticker VARCHAR(255) NOT NULL,
	end VARCHAR(255) NOT NULL,
	form VARCHAR(255) NOT NULL,
	val INTEGER,
	units VARCHAR(255),
	fy INTEGER,
	PRIMARY KEY (ticker,end,form)
);
`,
		}, `CollectionSetting`: {
			cols: []_refSqlTableCol{
				{`name`, `VARCHAR(255)`, ``, 0},
				{`value`, `TEXT`, `json`, 0},
			},
			create: `DROP TABLE IF EXISTS gensettings;
CREATE TABLE IF NOT EXISTS gensettings (
	name VARCHAR(255) NOT NULL,
	value TEXT,
	PRIMARY KEY (name)
);
`,
		}, `UiSetting`: {
			cols: []_refSqlTableCol{
				{`name`, `VARCHAR(255)`, ``, 0},
				{`value`, `TEXT`, `json`, 0},
			},
			create: `DROP TABLE IF EXISTS gensettings;
CREATE TABLE IF NOT EXISTS gensettings (
	name VARCHAR(255) NOT NULL,
	value TEXT,
	PRIMARY KEY (name)
);
`,
		},
		// End tabledefs
	}

	_refMetadatas = map[string]*_refMetadata{
		// Begin metadata
		`CollectionSetting`: {
			table:  "gensettings",
			tags:   []string{"name", "value"},
			fields: []string{"Name", "Value"},
			keys: map[string]*_refKeyMetadata{
				"": {
					tags:   []string{"name"},
					fields: []string{"Name"},
				},
			},
		}, `UiSetting`: {
			table:  "gensettings",
			tags:   []string{"name", "value"},
			fields: []string{"Name", "Value"},
			keys: map[string]*_refKeyMetadata{
				"": {
					tags:   []string{"name"},
					fields: []string{"Name"},
				},
			},
		}, `Company`: {
			table:  "gencompany",
			tags:   []string{"id", "name", "val", "fy"},
			fields: []string{"Id", "Name", "Value", "FoundedYear"},
			keys: map[string]*_refKeyMetadata{
				"": {
					tags:   []string{"id"},
					fields: []string{"Id"},
				},
				"b": {
					tags:   []string{"name"},
					fields: []string{"Name"},
				},
				"c": {
					tags:   []string{"fy"},
					fields: []string{"FoundedYear"},
				},
			},
		}, `Events`: {
			table:  "genevents",
			tags:   []string{"time", "name", "value"},
			fields: []string{"Time", "Name", "Value"},
			keys: map[string]*_refKeyMetadata{
				"": {
					tags:   []string{"time"},
					fields: []string{"Time"},
				},
			},
		}, `Filing`: {
			table:  "genfiling",
			tags:   []string{"ticker", "end", "form", "val", "units", "fy"},
			fields: []string{"Ticker", "EndDate", "Form", "Value", "Units", "FiscalYear"},
			keys: map[string]*_refKeyMetadata{
				"": {
					tags:   []string{"ticker", "end", "form"},
					fields: []string{"Ticker", "EndDate", "Form"},
				},
			},
		},
		// End metadata
	}
)
