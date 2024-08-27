package enc

type Flags int

const (
	// Indicates a key with an autoincrement value. Autoinc
	// globals have a unique key across the table.
	FlagAutoIncGlobal Flags = 1 << iota
	// Local autoincs only apply to storage layers that have
	// some notion of nesting or folders. In this case the
	// key will only be unique to the containing folder.
	FlagAutoIncLocal
	FlagEnd = iota
	// Clients that want to add flags should use:
	// 	AnotherFlag enc.Flags = 1 << (iota + enc.FlagEnd)
)
