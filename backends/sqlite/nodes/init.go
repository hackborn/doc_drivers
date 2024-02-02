package nodes

import (
	"embed"

	"github.com/hackborn/onefunc/pipeline"
)

func RegisterNodes() {
	pipeline.RegisterNode("go", func() pipeline.Node {
		return newGoNode()
	})
	pipeline.RegisterNode("sql", func() pipeline.Node {
		return newSqlNode()
	})
}

//go:embed templates
var templatesFs embed.FS
