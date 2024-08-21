package bboltrefdriver

import (
	"encoding/json"
	"sync/atomic"

	"github.com/hackborn/onefunc/values"
)

type _refMetadataNewConvFunc func() any

type _refMetadata struct {
	rootBucket    string
	buckets       []_refKeyMetadata
	newConvStruct _refMetadataNewConvFunc

	dk atomic.Pointer[[]string] // List of the buckets/domainNames
}

// toDb converts a domain value for this metadata into a database
// value. Database values are just copies of the domain value with
// metadata appropriate for the JSON schema for the database.
func (m *_refMetadata) toDb(src any) (any, error) {
	dst := m.newConvStruct()
	err := values.Copy(dst, src)
	return dst, err
}

// fromDb reads raw database data into a domain struct.
func (m *_refMetadata) fromDb(dst any, dbdata []byte) (any, error) {
	src := m.newConvStruct()
	err := json.Unmarshal(dbdata, src)
	if err != nil {
		return nil, err
	}
	err = values.Copy(dst, src)
	return dst, err
}

func (m *_refMetadata) DomainKeys() []string {
	if p := m.dk.Load(); p != nil {
		return *p
	}
	p := _refMakeDomainNames(m.buckets)
	m.dk.Store(&p)
	return p
}

type _refKeyMetadata struct {
	// domainName is the name of the field struct.
	domainName string

	// boltName is the name used for bbolt. It will either be
	// the name assigned in a doc tag, the domainName, or the
	// domainName modified in some way, i.e. through the lowercase flag.
	boltName string

	// Data type of this key.
	ft fieldType

	// leaf indicates this should be the key used to store the value,
	// instead of a bucket. If leaf is false then the key will be a
	// composite of all prior keys.
	leaf bool

	// autoInc indicates this is an automatically incrementing key.
	autoInc bool
}

func _refMakeDomainNames(keys []_refKeyMetadata) []string {
	c := make([]string, len(keys))
	for i, k := range keys {
		c[i] = k.domainName
	}
	return c
}
