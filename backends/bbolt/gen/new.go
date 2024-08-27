package bboltgendriver

import (
	"github.com/hackborn/doc"
)

func NewDriver(name string) doc.Driver {
	return &genDriver{}
}
