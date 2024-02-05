package nodes

const (
	FormatSqlite   = "sqlite"
	TemplateFsName = FormatSqlite + "templates"

	definitionKey = "def"

	templateDatestampKey   = "{{.Datestamp}}"
	templatePrefixKey      = "{{.Prefix}}"
	templatePackageKey     = "{{.Package}}"
	templateUtilPackageKey = "{{.UtilPackage}}"
)
