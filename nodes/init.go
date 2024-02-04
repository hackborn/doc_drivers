package nodes

import (
	"github.com/hackborn/onefunc/pipeline"
)

func init() {
	pipeline.RegisterNode("compare", func() pipeline.Runner {
		return &compareReportsNode{}
	})
	pipeline.RegisterNode("rungen", func() pipeline.Runner {
		return newRunDocDriverNode("gen")
	})
	pipeline.RegisterNode("runref", func() pipeline.Runner {
		return newRunDocDriverNode("ref")
	})
}
