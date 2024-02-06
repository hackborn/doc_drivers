package sqliterefdriver

import (
	"sync/atomic"
)

type _refMetadata struct {
	table  string   // table name
	tags   []string // List of all the tags in the table
	fields []string // List of all field names in the source struct, corresponding to tags
	keys   map[string]*_refKeyMetadata

	ttf atomic.Pointer[map[string]string] // Map of tags to their assocated fields. Generated value, only call TagsToFields().
	ftt atomic.Pointer[map[string]string] // Map of fields to their assocated tags. Generated value, only call FieldsToTags().
}

func (m *_refMetadata) TagsToFields() map[string]string {
	if p := m.ttf.Load(); p != nil {
		return *p
	}
	p := _refMakeMetadataMap(m.tags, m.fields)
	m.ttf.Store(&p)
	return p
}

func (m *_refMetadata) FieldsToTags() map[string]string {
	if p := m.ftt.Load(); p != nil {
		return *p
	}
	p := _refMakeMetadataMap(m.fields, m.tags)
	m.ftt.Store(&p)
	return p
}

type _refKeyMetadata struct {
	tags   []string
	fields []string

	ftt atomic.Pointer[map[string]string] // Map of fields to their assocated tags. Generated value, only call FieldsToTags().
}

func (m *_refKeyMetadata) FieldsToTags() map[string]string {
	if p := m.ftt.Load(); p != nil {
		return *p
	}
	p := _refMakeMetadataMap(m.fields, m.tags)
	m.ftt.Store(&p)
	return p
}

func _refMakeMetadataMap(keys, values []string) map[string]string {
	c := make(map[string]string)
	for i, k := range keys {
		c[k] = values[i]
	}
	return c
}
