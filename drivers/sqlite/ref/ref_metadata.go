package sqliterefdriver

import (
	"sync/atomic"
)

type _toxMetadata struct {
	table  string   // table name
	tags   []string // List of all the tags in the table
	fields []string // List of all field names in the source struct, corresponding to tags
	keys   map[string]*_toxKeyMetadata

	ttf atomic.Pointer[map[string]string] // Map of tags to their assocated fields. Generated value, only call TagsToFields().
	ftt atomic.Pointer[map[string]string] // Map of fields to their assocated tags. Generated value, only call FieldsToTags().
}

func (m *_toxMetadata) TagsToFields() map[string]string {
	if p := m.ttf.Load(); p != nil {
		return *p
	}
	p := _toxMakeMetadataMap(m.tags, m.fields)
	m.ttf.Store(&p)
	return p
}

func (m *_toxMetadata) FieldsToTags() map[string]string {
	if p := m.ftt.Load(); p != nil {
		return *p
	}
	p := _toxMakeMetadataMap(m.fields, m.tags)
	m.ftt.Store(&p)
	return p
}

type _toxKeyMetadata struct {
	tags   []string
	fields []string

	ftt atomic.Pointer[map[string]string] // Map of fields to their assocated tags. Generated value, only call FieldsToTags().
}

func (m *_toxKeyMetadata) FieldsToTags() map[string]string {
	if p := m.ftt.Load(); p != nil {
		return *p
	}
	p := _toxMakeMetadataMap(m.fields, m.tags)
	m.ftt.Store(&p)
	return p
}

func _toxMakeMetadataMap(keys, values []string) map[string]string {
	c := make(map[string]string)
	for i, k := range keys {
		c[k] = values[i]
	}
	return c
}

var (
	_toxMetadatas = map[string]*_toxMetadata{
		"Company": &_toxMetadata{
			table:  "Company",
			tags:   []string{"id", "name", "val", "fy"},
			fields: []string{"Id", "Name", "Value", "FoundedYear"},
			keys: map[string]*_toxKeyMetadata{
				"": &_toxKeyMetadata{
					tags:   []string{"id"},
					fields: []string{"Id"},
				},
				"b": &_toxKeyMetadata{
					tags:   []string{"name"},
					fields: []string{"Name"},
				},
			},
		},
		"Filing": &_toxMetadata{
			table:  "Filing",
			tags:   []string{"ticker", "end", "form", "val", "units", "fy"},
			fields: []string{"Ticker", "EndDate", "Form", "Value", "Units", "FiscalYear"},
			keys: map[string]*_toxKeyMetadata{
				"": &_toxKeyMetadata{
					tags:   []string{"ticker", "end", "form"},
					fields: []string{"Ticker", "EndDate", "Form"},
				},
			},
		},
	}
)
