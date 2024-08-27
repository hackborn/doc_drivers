package bboltrefdriver

type boltKey = []byte

var _refKeySep = []byte("/")

type fieldType uint8

const (
	stringType fieldType = iota
	uint64Type
)

// Copied from enc/
type keyFlags int

const (
	FlagAutoIncGlobal keyFlags = 1 << iota
	FlagAutoIncLocal
)

var (
	_refMetadatas = map[string]*_refMetadata{
		// Begin metadata
		`CollectionSetting`: {
			rootBucket: "settings",
			buckets: []_refKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: true, flags: 0},
			},
			newConvStruct: func() any { return &_refJsonCollectionSetting{} },
		},
		`Company`: {
			rootBucket: "company",
			buckets: []_refKeyMetadata{
				{domainName: "Id", boltName: "id", ft: stringType, leaf: false, flags: 0},
				{domainName: "Name", boltName: "name", ft: stringType, leaf: false, flags: 0},
				{domainName: "FoundedYear", boltName: "fy", ft: stringType, leaf: false, flags: 0},
			},
			newConvStruct: func() any { return &_refJsonCompany{} },
		},
		`Events`: {
			rootBucket: "events",
			buckets: []_refKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: false, flags: 0},
				{domainName: "Time", boltName: "time", ft: uint64Type, leaf: true, flags: 1},
			},
			newConvStruct: func() any { return &_refJsonEvents{} },
		},
		`Filing`: {
			rootBucket: "filing",
			buckets: []_refKeyMetadata{
				{domainName: "Ticker", boltName: "ticker", ft: stringType, leaf: false, flags: 0},
				{domainName: "EndDate", boltName: "end", ft: stringType, leaf: false, flags: 0},
				{domainName: "Form", boltName: "form", ft: stringType, leaf: false, flags: 0},
			},
			newConvStruct: func() any { return &_refJsonFiling{} },
		},
		`UiSetting`: {
			rootBucket: "settings",
			buckets: []_refKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: true, flags: 0},
			},
			newConvStruct: func() any { return &_refJsonUiSetting{} },
		},

		// End metadata
	}
)
