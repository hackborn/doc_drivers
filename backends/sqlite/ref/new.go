package sqliterefdriver

import (
	"github.com/hackborn/doc"
)

func NewDriver(name string) doc.Driver {
	return &_toxDriver{sqlDriverName: name}
}
