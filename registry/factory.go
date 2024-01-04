package registry

type Factory struct {
	Name       string
	DriverName string
	DbPath     string

	// ReferenceFiles is any source code files that will be used as
	// references to create the tox template files.
	ReferenceFiles map[string]string

	// New is a function to generate a new driver instance.
	New NewDriverFunc

	// Clients can add additional processing when generating templates.
	ProcessTemplate ProcessTemplateFunc
}
