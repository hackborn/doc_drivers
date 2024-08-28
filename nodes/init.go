package nodes

import (
	"embed"

	"github.com/hackborn/onefunc/pipeline"
)

func init() {
	// Register the test data
	pipeline.RegisterFs("testnodedata", testNodeDataFs)

	pipeline.RegisterNode("compare", func() pipeline.Node {
		return &compareReportsNode{}
	})
	pipeline.RegisterNode("testgen", func() pipeline.Node {
		return newTestDocDriverNode("gen")
	})
	pipeline.RegisterNode("testref", func() pipeline.Node {
		return newTestDocDriverNode("ref")
	})
	pipeline.RegisterNode("rungen", func() pipeline.Node {
		return newRunDocDriverNode("gen")
	})
	pipeline.RegisterNode("runref", func() pipeline.Node {
		return newRunDocDriverNode("ref")
	})
}

//go:embed testnodedata/*
var testNodeDataFs embed.FS
