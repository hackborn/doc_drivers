package main

import (
	"fmt"

	"github.com/hackborn/doc_drivers"
)

// main generates the reference templates.
func main() {
	f, err := drivers.GetFactoryFromCla()
	if err != nil {
		fmt.Println(err)
		fmt.Println(help)
		return
	}
	m := newMakeTemplates(WithPrefix("_tox"))
	m.Run(f.ReferenceFiles)
}
