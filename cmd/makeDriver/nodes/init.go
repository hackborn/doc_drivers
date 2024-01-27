package nodes

import (
	"embed"

	"github.com/hackborn/onefunc/pipeline"
)

func init() {
	pipeline.RegisterNode("go", func() pipeline.Node {
		return newGoNode()
	})
	pipeline.RegisterNode("sql", func() pipeline.Node {
		return newSqlNode()
	})
}

//go:embed templates
var templates embed.FS
