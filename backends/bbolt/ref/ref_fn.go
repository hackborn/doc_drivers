package bboltrefdriver

import (
	"fmt"

	"github.com/hackborn/doc"
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

func newGetBucketValuesHandler(rootBucket string, buckets []_refKeyMetadata) *getBucketValuesHandler {
	values := make([]string, len(buckets)+1)
	values[0] = rootBucket
	return &getBucketValuesHandler{buckets: buckets,
		values: values,
	}
}

type getBucketValuesHandler struct {
	buckets []_refKeyMetadata
	// Values is all the values in the buckets, with the 0 value set to
	// the root bucket, so the size will be len(buckets)+1.
	values []string
}

// makeKey combines the values into a single key, reporting an
// error if anyone is missing.
func (h *getBucketValuesHandler) makeKey() (string, error) {
	key := ""
	for i, s := range h.values {
		if s == "" {
			return "", fmt.Errorf("Missing value for %v", h.buckets[i].domainName)
		}
		if i > 0 {
			key += "/"
		}
		key += s
	}
	return key, nil
}

func (h *getBucketValuesHandler) Handle(name string, value any) (string, any) {
	for i, b := range h.buckets {
		if b.domainName == name {
			if s, ok := value.(string); ok {
				// +1 skips the root bucket
				h.values[i+1] = s
			}
		}
	}
	return "", nil
}
