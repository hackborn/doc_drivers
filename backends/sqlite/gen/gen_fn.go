package sqlitegendriver

// autogenerated with github.com/hackborn/doc_drivers on 2024-05-20
// do not modify

import (
	"cmp"
	"encoding/json"
	"fmt"

	"github.com/hackborn/doc"
	"github.com/hackborn/onefunc/errors"
	ofstrings "github.com/hackborn/onefunc/strings"
)

// genNewFormat answers a new Format for the driver,
// containing the formatting rules for translating
// expressions.
func genNewFormat() doc.Format {
	keywords := map[string]string{
		doc.AndKeyword:    " AND ",
		doc.AssignKeyword: " = ",
		doc.OrKeyword:     " OR ",
	}
	return &genFormat{keywords: keywords}
}

type genFormat struct {
	keywords map[string]string
}

func (f *genFormat) Keyword(s string) string {
	s, _ = f.keywords[s]
	return s
}

func (f *genFormat) Value(v interface{}) (string, error) {
	s := fmt.Sprintf("%v", v)
	switch v.(type) {
	case string:
		c := string('\'')
		s = c + s + c
	}
	return s, nil
}

type fieldsAndValuesHandler struct {
	err    error
	fields []any
	values []any
	cols   []genSqlTableCol
}

func (h *fieldsAndValuesHandler) Handle(name string, value any) (string, any) {
	h.fields = append(h.fields, name)
	//	h.values = append(h.values, value)
	formatted := h.formatValue(name, value)
	h.values = append(h.values, formatted)
	return name, value
}

// formatValue applies any desired formmating to the value.
func (h *fieldsAndValuesHandler) formatValue(name string, value any) any {
	format := h.formatForTag(name)
	switch format {
	case "json":
		if dat, err := json.Marshal(value); err == nil {
			return string(dat)
		} else {
			h.err = cmp.Or(h.err, err)
			return value
		}
	default:
		return value
	}
}

func (h *fieldsAndValuesHandler) formatForTag(tag string) string {
	for _, v := range h.cols {
		if v.name == tag {
			return v.format
		}
	}
	return ""
}

func makeExcludedFieldValues(eb errors.Block, names []any) string {
	w := ofstrings.GetWriter(eb)
	defer ofstrings.PutWriter(w)

	for i, n := range names {
		if i > 0 {
			w.WriteString(", ")
		}
		name := fmt.Sprintf("%v", n)
		w.WriteString(name + "=excluded." + name)
	}
	return ofstrings.String(w)
}

func getFormats(fields []string) []string {
	s := make([]string, 0, len(fields))
	return s
}

func whereClause(req doc.GetRequest) (string, error) {
	if req.Condition == nil {
		return "", nil
	}
	expr, err := req.Condition.Compile()
	if err != nil {
		return "", err
	}
	s, err := expr.Format()
	if err != nil {
		return "", err
	}
	if s == "" {
		return "", nil
	}
	return " WHERE " + s, nil
}

// getTagsAndFields extracts the tag names from the request and returns
// a slice of the tags with their associated fields.
func getTagsAndFields(meta *genMetadata, req doc.GetRequest) ([]string, []string) {
	if req.Fields == nil {
		return meta.tags, meta.fields
	}
	tags := req.Fields.Fields()
	if len(tags) < 1 {
		return meta.tags, meta.fields
	}
	// Is there a way to cache this?
	fields := make([]string, 0, len(tags))
	ttf := meta.TagsToFields()
	for _, t := range tags {
		if f, ok := ttf[t]; ok {
			fields = append(fields, f)
		} else {
			fields = append(fields, t)
		}
	}
	return tags, fields
}
