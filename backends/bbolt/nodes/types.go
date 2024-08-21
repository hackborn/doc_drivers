package nodes

import (
	"cmp"
	"fmt"
	"strings"
)

// MetadataDef is used by the  const template file.
type MetadataDef struct {
	DomainName    string
	RootBucket    string
	Buckets       []MetadataKeyDef
	NewConvStruct string
}

func (m MetadataDef) Validate() error {
	if m.RootBucket == "" {
		return fmt.Errorf("Metadata must have a root bucket")
	}
	if len(m.Buckets) < 1 {
		return fmt.Errorf("Metadata must have at least one key tag")
	}
	return nil
}

type MetadataKeyDef struct {
	DomainName string
	BoltName   string
	Ft         string
	AutoInc    bool
	Leaf       bool
	keyInfo    *metadataKeyInfo
}

// metadataKeyInfo is used during parsing to sort the keys.
type metadataKeyInfo struct {
	group string
	index int
}

func compareKeys(a, b *metadataKeyInfo) int {
	if a == nil && b == nil {
		return 0
	} else if a == nil {
		return 1
	} else if b == nil {
		return -1
	}
	if a.group == b.group {
		return cmp.Compare(a.index, b.index)
	}
	return strings.Compare(a.group, b.group)
}

// JsonDef is used by the json template file.
type JsonDef struct {
	Name   string
	Fields []JsonFieldDef
}

type JsonFieldDef struct {
	Name string
	Type string
	Tag  string
}
