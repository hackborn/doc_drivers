package nodes

import (
	"github.com/hackborn/onefunc/pipeline"
)

func RegisterNodes() {
	pipeline.RegisterNode("go", func() pipeline.Node {
		return newGoNode()
	})
}
