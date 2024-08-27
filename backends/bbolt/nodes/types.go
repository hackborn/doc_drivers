package nodes

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/hackborn/doc_drivers/enc"
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
		return fmt.Errorf("Metadata for \"%v\" must have a root bucket", m.DomainName)
	}
	if len(m.Buckets) < 1 {
		return fmt.Errorf("Metadata for \"%v\" must have at least one key tag", m.DomainName)
	}
	// Can only have 1 autoinc key
	autoincCount := 0
	for _, key := range m.Buckets {
		if key.IsAutoInc() {
			autoincCount++
			if autoincCount > 1 {
				return fmt.Errorf("Metadata for \"%v\" must not have more than one autoinc tag", m.DomainName)
			}
		}
	}
	return nil
}

// sortAutoInc places the autoinc tag at the tail;
func (m MetadataDef) sortAutoInc() {
	var autoinc *MetadataKeyDef
	writeI := 0
	for i, key := range m.Buckets {
		if key.IsAutoInc() {
			autoinc = &key
			continue
		}

		if writeI != i {
			m.Buckets[writeI] = key
		}
		writeI++
	}
	if autoinc != nil {
		m.Buckets[len(m.Buckets)-1] = *autoinc
	}
}

// setLeaf sets the def to a leaf for certain cases.
func (m *MetadataDef) setLeaf() {
	// Set the leaf value here. Keys are a leaf if they
	// are the only key, or they are the final key and they
	// auto increment.
	if len(m.Buckets) == 1 {
		m.Buckets[0].Leaf = true
	} else if len(m.Buckets) > 1 && m.Buckets[len(m.Buckets)-1].IsAutoInc() {
		m.Buckets[len(m.Buckets)-1].Leaf = true
	}
}

type MetadataKeyDef struct {
	DomainName string
	BoltName   string
	Ft         string
	Flags      enc.Flags
	Leaf       bool
	keyInfo    *metadataKeyInfo
}

func (d MetadataKeyDef) IsAutoInc() bool {
	return d.Flags&enc.FlagAutoIncGlobal != 0 || d.Flags&enc.FlagAutoIncLocal != 0
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
