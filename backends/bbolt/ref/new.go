package bboltrefdriver

import (
	"github.com/hackborn/doc"
)

func NewDriver(name string) doc.Driver {
	return &_refDriver{}
}
