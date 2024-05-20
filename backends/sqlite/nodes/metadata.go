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

// PrimaryKey answers the "primary" key in the key map. This is
// just defined as the first key in alphabetic order.
func (d metadata) PrimaryKey() (string, bool) {
	if len(d.Keys) < 1 {
		return "", false
	}
	first := true
	key := ""
	for k, _ := range d.Keys {
		if first {
			key = k
			first = false
		} else {
			if strings.Compare(k, key) < 0 {
				key = k
			}
		}
	}
	return key, true
}

type structField struct {
	Tag    string
	Field  string
	Type   string
	Format string
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
