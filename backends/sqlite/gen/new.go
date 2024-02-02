package sqlitegendriver

import (
	"github.com/hackborn/doc"
)

func NewDriver(name string) doc.Driver {
	return &genDriver{sqlDriverName: name}
}
