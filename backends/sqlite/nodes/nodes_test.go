package nodes

import (
	"testing"

	"github.com/hackborn/onefunc/jacl"
	"github.com/hackborn/onefunc/pipeline"
)

// ---------------------------------------------------------
// TEST-METADATA
func TestMetadata(t *testing.T) {
	table := []struct {
		structData *pipeline.StructData
		cmp        []string
		wantErr    error
	}{
		{companyStruct, []string{`Name=Company`, `Fields/0/Tag=id1`, `Fields/0/Field=Id1`, `Fields/1/Tag=id2a`}, nil},
		{companyStruct, []string{`Keys/""/0/Tag=id1`}, nil},
		{companyStruct, []string{`Keys/a/0/Tag=id2a`, `Keys/a/1/Tag=id2b`}, nil},
		{companyStruct, []string{`Keys/b/0/Tag=id3b`, `Keys/b/1/Tag=id3c`, `Keys/b/2/Tag=id3a`}, nil},
	}
	for i, v := range table {
		md, haveErr := makeMetadata(v.structData)
		cmpErr := jacl.Run(md, v.cmp...)

		if v.wantErr == nil && haveErr != nil {
			t.Fatalf("TestMetadata %v expected no error but has %v", i, haveErr)
		} else if v.wantErr != nil && haveErr == nil {
			t.Fatalf("TestMetadata %v has no error but exptected %v", i, v.wantErr)
		} else if cmpErr != nil {
			t.Fatalf("TestMetadata %v comparison error: %v", i, cmpErr)
		}
	}
}

// ---------------------------------------------------------
// TEST-DATA

var (
	companyStruct = &pipeline.StructData{
		Name: "Company",
		Fields: []pipeline.StructField{
			{Name: "Id1", Tag: "id1, key"},
			{Name: "Id2a", Tag: "id2a, key(a)"},
			{Name: "Id2b", Tag: "id2b, key (a)"},
			{Name: "Id3a", Tag: "id3a, key(b, 2)"},
			{Name: "Id3b", Tag: "id3b, key(b, 0)"},
			{Name: "Id3c", Tag: "id3c, key(b,1)"},
			{Name: "Name", Tag: "name"},
		},
	}
)
