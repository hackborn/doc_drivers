package nodes

import (
	"fmt"
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
		wantCmpErr error
	}{
		{keyStruct, []string{`Name=Key`, `Fields/0/Tag=id1`, `Fields/0/Field=Id1`, `Fields/1/Tag=id2a`}, nil, nil},
		{keyStruct, []string{`Keys/""/0/Tag=id1`}, nil, nil},
		{keyStruct, []string{`Keys/a/0/Tag=id2a`, `Keys/a/1/Tag=id2b`}, nil, nil},
		{keyStruct, []string{`Keys/b/0/Tag=id3b`, `Keys/b/1/Tag=id3c`, `Keys/b/2/Tag=id3a`}, nil, nil},
		{nameStruct, []string{`Name=Name`, `Fields/0/Field=Name`, `Fields/0/Tag=name`}, nil, nil},
		{nameStruct, []string{`Fields/1/Field=FieldName1`, `Fields/1/Tag=fieldname1`}, nil, nil},
		{nameStruct, []string{`Fields/2/Field=FieldName2`, `Fields/2/Tag=fieldname2`}, nil, nil},
		{nameStruct, []string{`Fields/3/Field=KeyName1`, `Fields/3/Tag=keyname1`}, nil, nil},
		{nameStruct, []string{`Keys/""/0/Field=KeyName1`, `Keys/""/0/Tag=keyname1`}, nil, nil},
		{skipStruct, []string{`Fields/0/Field=Skip1`}, nil, fmt.Errorf("out-of-range because skip fields don't exist")},
	}
	for i, v := range table {
		md, _, haveErr := makeMetadata(v.structData, "")
		cmpErr := jacl.Run(md, v.cmp...)

		if v.wantErr == nil && haveErr != nil {
			t.Fatalf("TestMetadata %v expected no error but has %v", i, haveErr)
		} else if v.wantErr != nil && haveErr == nil {
			t.Fatalf("TestMetadata %v has no error but exptected %v", i, v.wantErr)
		} else if cmpErr != nil && v.wantCmpErr == nil {
			t.Fatalf("TestMetadata %v comparison error: %v", i, cmpErr)
		}
	}
}

// ---------------------------------------------------------
// TEST-DATA

var (
	keyStruct = &pipeline.StructData{
		Name: "Key",
		Fields: []pipeline.StructField{
			{Name: "Id1", Tag: "name(id1), key"},
			{Name: "Id2a", Tag: "name(id2a), key(a)"},
			{Name: "Id2b", Tag: "name(id2b), key (a)"},
			{Name: "Id3a", Tag: "name(id3a), key(b, 2)"},
			{Name: "Id3b", Tag: "name(id3b), key(b, 0)"},
			{Name: "Id3c", Tag: "name(id3c), key(b,1)"}, // 5
		},
	}

	nameStruct = &pipeline.StructData{
		Name: "Name",
		Fields: []pipeline.StructField{
			{Name: "Name", Tag: "name(name)"},
			{Name: "FieldName1", Tag: ""},
			{Name: "FieldName2", Tag: ","},
			{Name: "KeyName1", Tag: "key"},
		},
	}

	skipStruct = &pipeline.StructData{
		Name: "Skip",
		Fields: []pipeline.StructField{
			{Name: "Skip1", Tag: "-"},
			{Name: "Skip2", Tag: "-,"},
		},
	}
)
