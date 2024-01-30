package sqlitegendriver

// autogenerated with github.com/hackborn/onefunc/pipeline on 29 Jan 24 16:52 PST
// do not modify

import (
	"sync/atomic"
)

type genMetadata struct {
	table  string   // table name
	tags   []string // List of all the tags in the table
	fields []string // List of all field names in the source struct, corresponding to tags
	keys   map[string]*genKeyMetadata

	ttf atomic.Pointer[map[string]string] // Map of tags to their assocated fields. Generated value, only call TagsToFields().
	ftt atomic.Pointer[map[string]string] // Map of fields to their assocated tags. Generated value, only call FieldsToTags().
}

func (m *genMetadata) TagsToFields() map[string]string {
	if p := m.ttf.Load(); p != nil {
		return *p
	}
	p := genMakeMetadataMap(m.tags, m.fields)
	m.ttf.Store(&p)
	return p
}

func (m *genMetadata) FieldsToTags() map[string]string {
	if p := m.ftt.Load(); p != nil {
		return *p
	}
	p := genMakeMetadataMap(m.fields, m.tags)
	m.ftt.Store(&p)
	return p
}

type genKeyMetadata struct {
	tags   []string
	fields []string

	ftt atomic.Pointer[map[string]string] // Map of fields to their assocated tags. Generated value, only call FieldsToTags().
}

func (m *genKeyMetadata) FieldsToTags() map[string]string {
	if p := m.ftt.Load(); p != nil {
		return *p
	}
	p := genMakeMetadataMap(m.fields, m.tags)
	m.ftt.Store(&p)
	return p
}

func genMakeMetadataMap(keys, values []string) map[string]string {
	c := make(map[string]string)
	for i, k := range keys {
		c[k] = values[i]
	}
	return c
}

var (
	genMetadatas = map[string]*genMetadata{
		"Company": &genMetadata{
			table:  "Company",
			tags:   []string{"id", "name", "val", "fy"},
			fields: []string{"Id", "Name", "Value", "FoundedYear"},
			keys: map[string]*genKeyMetadata{
				"": &genKeyMetadata{
					tags:   []string{"id"},
					fields: []string{"Id"},
				},
				"b": &genKeyMetadata{
					tags:   []string{"name"},
					fields: []string{"Name"},
				},
			},
		},
		"Filing": &genMetadata{
			table:  "Filing",
			tags:   []string{"ticker", "end", "form", "val", "units", "fy"},
			fields: []string{"Ticker", "EndDate", "Form", "Value", "Units", "FiscalYear"},
			keys: map[string]*genKeyMetadata{
				"": &genKeyMetadata{
					tags:   []string{"ticker", "end", "form"},
					fields: []string{"Ticker", "EndDate", "Form"},
				},
			},
		},
	}
)