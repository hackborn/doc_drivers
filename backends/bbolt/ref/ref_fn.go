package bboltrefdriver

import (
	"encoding/binary"
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

// _refToBoltKey converts values into []byte values used as bolt keys.
func _refToBoltKey(value any) (boltKey, bool) {
	switch t := value.(type) {
	case uint64:
		return _refItob(t), true
	case string:
		return []byte(t), true
	}
	return nil, false
}

// _refItob returns an 8-byte big endian representation of v.
func _refItob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
