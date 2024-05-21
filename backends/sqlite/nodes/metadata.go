package nodes

import (
	"slices"
	"strings"

	oferrors "github.com/hackborn/onefunc/errors"
	"github.com/hackborn/onefunc/pipeline"
	ofslices "github.com/hackborn/onefunc/slices"
)

// metadata is parallel to pipeline.StructData, except with parsed tags.
type metadata struct {
	// The table name
	Name   string
	Fields []structField
	Keys   map[string][]structKey
}

func (d metadata) TagNames() []string {
	return ofslices.ArrayFrom(d.Fields, func(f structField) string {
		return f.Tag
	})
}

func (d metadata) FieldNames() []string {
	return ofslices.ArrayFrom(d.Fields, func(f structField) string {
		return f.Field
	})
}

func (d metadata) KeyTagNames(key string) []string {
	return ofslices.ArrayFrom(d.Keys[key], func(key structKey) string {
		return key.Tag
	})
}

func (d metadata) KeyFieldNames(key string) []string {
	return ofslices.ArrayFrom(d.Keys[key], func(key structKey) string {
		return key.Field
	})
}

func (d metadata) KeySpecs() keySpecList {
	list := keySpecList{}
	for name, v := range d.Keys {
		groupSpec := keyGroupSpec{name: name}
		for _, key := range v {
			keySpec := keySpec{Field: key.Field, ColumnName: key.Tag}
			if field, ok := d.fieldForTag(key.Tag); ok {
				keySpec.DbType = convertGoTypeToSQLType(field.Type)
			}
			groupSpec.keys = append(groupSpec.keys, keySpec)
		}
		list.keyGroups = append(list.keyGroups, groupSpec)
	}
	// Sort so primary key is first
	slices.SortFunc(list.keyGroups, func(a, b keyGroupSpec) int {
		return strings.Compare(a.name, b.name)
	})
	return list
}

func (d metadata) fieldForTag(tag string) (structField, bool) {
	for _, sf := range d.Fields {
		if sf.Tag == tag {
			return sf, true
		}
	}
	return structField{}, false
}

type structField struct {
	Tag    string
	Field  string
	Type   string
	Format string // A format to translate to when storing in the database.
}

type structKey struct {
	Tag   string
	Field string
}

type parsedKey struct {
	name     string
	position int

	tagName   string
	fieldName string
}

// makeMetadata answers the results of parsing the struct
// data, including the tags, into a parallel structure.
func makeMetadata(pin *pipeline.StructData, tablePrefix string) (metadata, error) {
	eb := oferrors.FirstBlock{}
	md := metadata{Name: pin.Name}
	md.Keys = make(map[string][]structKey)
	keys := make(map[string][]*parsedKey)
	for _, f := range pin.Fields {
		pt, err := parseTag(f.Tag)
		eb.AddError(err)
		sf, pk := convertToLocal(f, pt)
		// Skip indicator
		if sf.Tag == "-" {
			continue
		}
		// Default field name indicator.
		if sf.Tag == "" {
			// SQLITE convention is lowercase names
			sf.Tag = strings.ToLower(sf.Field)
		}
		md.Fields = append(md.Fields, sf)
		if pk != nil {
			pk.tagName = sf.Tag
			pk.fieldName = sf.Field
			if found, ok := keys[pk.name]; ok {
				found = append(found, pk)
				keys[pk.name] = found
			} else {
				keys[pk.name] = []*parsedKey{pk}
			}
		}
	}
	// Compile the keys
	for k, v := range keys {
		slices.SortFunc(v, func(a, b *parsedKey) int {
			if a.position < b.position {
				return -1
			} else if a.position > b.position {
				return 1
			} else {
				return 0
			}
		})
		value := make([]structKey, 0, len(v))
		for _, vv := range v {
			value = append(value, structKey{Tag: vv.tagName, Field: vv.fieldName})
		}
		md.Keys[k] = value
	}
	makeTableMetadata(pin, &md, &eb)
	md.Name = tablePrefix + md.Name
	return md, eb.Err
}

func makeTableMetadata(pin *pipeline.StructData, md *metadata, eb oferrors.Block) {
	for _, f := range pin.UnexportedFields {
		// The tag was filtered for my "doc" keyword, so any non-empty
		// tag will be a table specification
		if f.Tag == "" {
			continue
		}
		pt, err := parseTag(f.Tag)
		eb.AddError(err)
		if pt.name != "" {
			md.Name = pt.name
		}
	}
}

// convertToLocal converts a parsed tag to struct field and parsed key.
func convertToLocal(f pipeline.StructField, parsed parsedTag) (structField, *parsedKey) {
	sf := structField{Tag: parsed.name, Field: f.Name, Type: f.Type}
	var key *parsedKey
	if parsed.hasKey {
		key = &parsedKey{name: parsed.keyGroup, position: parsed.keyIndex}
	}
	return sf, key
}

// keySpecList is the ordered list of key metadata.
// Keys have a variety of representations in this
// metadata mess, this is targeted as being the
// "full" representation.
type keySpecList struct {
	// keys is an ordered list of keys. The first
	// is the "primary."
	keyGroups []keyGroupSpec
}

func (s keySpecList) isPrimary(tag string) bool {
	if len(s.keyGroups) < 1 {
		return false
	}
	for _, keySpec := range s.keyGroups[0].keys {
		if keySpec.ColumnName == tag {
			return true
		}
	}
	return false
}

type keyGroupSpec struct {
	name string
	keys []keySpec
}

func (s keyGroupSpec) columnNames() []string {
	return ofslices.ArrayFrom(s.keys, func(key keySpec) string {
		return key.ColumnName
	})
}

type keySpec struct {
	// The field name in the source struct.
	Field string

	// The column name in the databse.
	ColumnName string

	// The data type in the database.
	DbType string
}
