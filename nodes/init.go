package nodes

import (
	"github.com/hackborn/onefunc/pipeline"
)

func init() {
	pipeline.RegisterNode("runref", func() pipeline.Node {
		return newRunDocDriverNode("ref")
	})
	pipeline.RegisterNode("rungen", func() pipeline.Node {
		return newRunDocDriverNode("gen")
	})
}
