package sqlitedriver

// autogenerated with github.com/hackborn/tox
// do not modify

const (
	_toxKeysVar        = "$KEYS$"
	_toxKeyValuesVar   = "$KEYVALUES$"
	_toxFieldsVar      = "$FIELDS$"
	_toxTableVar       = "$TABLE$"
	_toxValuesVar      = "$VALUES$"
	_toxFieldValuesVar = "$FIELDVALUES$"

	_toxAndKeyword      = "AND"
	_toxAndKeywordWS    = " AND "
	_toxEqualsKeyword   = "="
	_toxEqualsKeywordWS = " = "

	_toxQuoteSz = string(rune('\''))

	_toxSetSql = `INSERT INTO $TABLE$ ($FIELDS$) VALUES($VALUES$) ON CONFLICT($KEYS$) DO UPDATE SET $FIELDVALUES$;`
	_toxDelSql = `DELETE FROM $TABLE$ WHERE ($KEYVALUES$);`
)

var (
	_toxDefinitions = map[string]string{
		`Company`: `DROP TABLE IF EXISTS Company;
CREATE TABLE IF NOT EXISTS Company (
	id VARCHAR(255),
	name VARCHAR(255),
	val INTEGER,
	fy INTEGER,
	PRIMARY KEY (id)
);`,
		`Filing`: `DROP TABLE IF EXISTS Filing;
CREATE TABLE IF NOT EXISTS Filing (
	ticker VARCHAR(255),
	end VARCHAR(255),
	form VARCHAR(255),
	val INTEGER,
	units VARCHAR(255),
	fy INTEGER,
	PRIMARY KEY (ticker, end, form)
);`,
	}
)
