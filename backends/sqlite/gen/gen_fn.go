package sqlitegendriver

// autogenerated with github.com/hackborn/doc_drivers on 05 Feb 24 22:20 PST
// do not modify

import (
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
	fields []any
	values []any
}

func (h *fieldsAndValuesHandler) Handle(name string, value any) (string, any) {
	h.fields = append(h.fields, name)
	h.values = append(h.values, value)
	return name, value
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
