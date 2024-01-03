package registry

import (
	"github.com/hackborn/doc"
)

// NewDriverFunc returns a new instance of a doc.Driver.
type NewDriverFunc func() doc.Driver

type ProcessTemplateFunc func()
