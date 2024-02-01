package nodes

import (
	"github.com/hackborn/onefunc/pipeline"
)

func init() {
	pipeline.RegisterNode("compare", func() pipeline.Node {
		return &compareReportsNode{}
	})
	pipeline.RegisterNode("rungen", func() pipeline.Node {
		return newRunDocDriverNode("gen")
	})
	pipeline.RegisterNode("runref", func() pipeline.Node {
		return newRunDocDriverNode("ref")
	})
}
