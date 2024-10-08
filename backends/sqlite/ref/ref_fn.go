package sqliterefdriver

import (
	"cmp"
	"encoding/json"
	"fmt"

	"github.com/hackborn/doc"
	"github.com/hackborn/onefunc/errors"
	ofstrings "github.com/hackborn/onefunc/strings"
)

// _refNewFormat answers a new Format for the driver,
// containing the formatting rules for translating
// expressions.
func _refNewFormat() doc.Format {
	keywords := map[string]string{
		doc.AndKeyword:    " AND ",
		doc.AssignKeyword: " = ",
		doc.OrKeyword:     " OR ",
	}
	return &_refFormat{keywords: keywords}
}

type _refFormat struct {
	keywords map[string]string
}

func (f *_refFormat) Keyword(s string) string {
	s, _ = f.keywords[s]
	return s
}

func (f *_refFormat) Value(v interface{}) (string, error) {
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
	cols   []_refSqlTableCol
	filter doc.Filter // Accept / reject adding to fields/values based on filter.
}

func (h *fieldsAndValuesHandler) Handle(name string, value any) (string, any) {
	col := getColByName(name, h.cols)
	if h.filter.Rule != 0 && !acceptCol(h.filter.Rule, col.name, col.flags) {
		return name, value
	}
	h.fields = append(h.fields, name)
	// Deal with any values that can't be stored directly in the DB by formatting them.
	formatted := h.formatValue(col, value)
	h.values = append(h.values, formatted)
	return name, value
}

// formatValue applies any desired formmating to the value.
func (h *fieldsAndValuesHandler) formatValue(col _refSqlTableCol, value any) any {
	switch col.format {
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

func getColByName(name string, cols []_refSqlTableCol) _refSqlTableCol {
	for _, c := range cols {
		if c.name == name {
			return c
		}
	}
	return _refSqlTableCol{}
}

func acceptCol(rule int64, name string, flags uint64) bool {
	switch rule {
	case 0, doc.RuleSetItem:
		return true
	case doc.RuleCreateItem:
		// Everything is true but auto inc
		return flags&colFlagAuto == 0
	default:
		return true
	}
}

// getTagsAndFields extracts the tag names from the request and returns
// a slice of the tags with their associated fields.
func getTagsAndFields(meta *_refMetadata, req doc.GetRequest) ([]string, []string) {
	if req.Fields == nil {
		return meta.tags, meta.fields
	}
	tags := req.Fields.Names()
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
