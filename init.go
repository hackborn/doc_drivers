package drivers

import (
	"reflect"

	_ "github.com/hackborn/doc_drivers/backends/bbolt"
	_ "github.com/hackborn/doc_drivers/backends/sqlite"
	_ "github.com/hackborn/doc_drivers/nodes"
	"github.com/hackborn/doc_drivers/registry"
	_ "github.com/hackborn/onefunc/pipeline/nodes"
)

func init() {
	type Empty struct {
	}
	registry.UtilPackageName = reflect.TypeOf(Empty{}).PkgPath()
}
