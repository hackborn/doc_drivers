package nodes

import (
	"embed"

	"github.com/hackborn/onefunc/pipeline"
)

func RegisterNodes() {
	pipeline.RegisterNode("go", func() pipeline.Runner {
		return newGoNode()
	})
	pipeline.RegisterNode("sql", func() pipeline.Runner {
		return newSqlNode()
	})
	pipeline.RegisterNode("template", func() pipeline.Runner {
		return newTemplateNode()
	})
}

//go:embed templates
var templatesFs embed.FS
