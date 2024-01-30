package registry

type Factory struct {
	Name       string
	DriverName string
	DbPath     string

	// ReferenceFiles is any source code files that will be used as
	// references to create the tox template files.
	ReferenceFiles map[string]string

	// NewRef is a function to generate a new doc driver instance based on the reference driver.
	NewRef NewDriverFunc

	// NewGenerated generates a new doc driver based on the generated driver.
	NewGenerated NewDriverFunc

	// Clients can add additional processing when generating templates.
	ProcessTemplate ProcessTemplateFunc
}
