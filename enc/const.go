package enc

type Flags int

const (
	FlagAutoIncGlobal Flags = 1 << iota
	FlagAutoIncLocal
	FlagEnd = iota
	// Clients that want to add flags should use:
	// 	AnotherFlag enc.Flags = 1 << (iota + enc.FlagEnd)
)
