package nodes

import (
	"github.com/hackborn/onefunc/pipeline"
)

type makeSqlPinFunc func(data *sqlNodeData, state *pipeline.State, pin *pipeline.StructData) (pipeline.Pin, error)

type makeTemplateContentFunc func(c *makeTemplateContent) error
