package nodes

import (
	"embed"

	"github.com/hackborn/onefunc/pipeline"
)

func init() {
	// Register the test data
	pipeline.RegisterFs("testnodedata", testNodeDataFs)

	pipeline.RegisterNode("testgen", func() pipeline.Node {
		return newTestDocDriverNode("gen")
	})
	pipeline.RegisterNode("testref", func() pipeline.Node {
		return newTestDocDriverNode("ref")
	})
}

//go:embed testnodedata/*
var testNodeDataFs embed.FS
