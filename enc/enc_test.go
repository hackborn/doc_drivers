package enc

import (
	"flag"
	"os"
	"testing"

	"github.com/hackborn/onefunc/jacl"
)

// go test -bench=.
// go test . -update
var (
	update = flag.Bool("update", false, "update the golden files of this test")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func isDrawing() bool {
	return update != nil && *update == true
}

// ---------------------------------------------------------
// TEST-PARSE-TAG
func TestParseTag(t *testing.T) {
	f := func(expr string, wantErr error, want ...string) {
		t.Helper()

		have, haveErr := ParseTag(expr)
		if err := jacl.RunErr(haveErr, wantErr); err != nil {
			t.Fatalf("Want err %v but have %v (%v)", wantErr, haveErr, err)
		} else if err := jacl.Run(have, want...); err != nil {
			t.Fatalf("Want %v but has %v (%v)", want, have, err)
		}
	}
	f("name(id)", nil, `Name=id`, `HasKey=false`)
	f("name(id), key", nil, `Name=id`, `HasKey=t`)
	f("name(id), key(a,1)", nil, `Name=id`, `HasKey=t`, `KeyGroup=a`, `KeyIndex=1`)
	f(`key, autoinc`, nil, `Name=""`, `HasKey=t`, `KeyGroup=""`, `KeyIndex=0`, `Flags=1`)
	f(`key, autoinc(local)`, nil, `Flags=2`)
	f(`format(json)`, nil, `Format=json`)
}
