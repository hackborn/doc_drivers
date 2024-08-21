package bboltrefdriver

type fieldType uint8

const (
	stringType fieldType = iota
	uint64Type
)

var (
	_refMetadatas = map[string]*_refMetadata{
		// Begin metadata
		`CollectionSetting`: {
			rootBucket: "settings",
			buckets: []_refKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: true, autoInc: false},
			},
			newConvStruct: func() any { return &_refJsonCollectionSetting{} },
		},
		`Company`: {
			rootBucket: "company",
			buckets: []_refKeyMetadata{
				{domainName: "Id", boltName: "id", ft: stringType, leaf: false, autoInc: false},
				{domainName: "Name", boltName: "name", ft: stringType, leaf: false, autoInc: false},
				{domainName: "FoundedYear", boltName: "fy", ft: stringType, leaf: false, autoInc: false},
			},
			newConvStruct: func() any { return &_refJsonCompany{} },
		},
		`Events`: {
			rootBucket: "events",
			buckets: []_refKeyMetadata{
				{domainName: "Time", boltName: "time", ft: uint64Type, leaf: true, autoInc: true},
			},
			newConvStruct: func() any { return &_refJsonEvents{} },
		},
		`Filing`: {
			rootBucket: "filing",
			buckets: []_refKeyMetadata{
				{domainName: "Ticker", boltName: "ticker", ft: stringType, leaf: false, autoInc: false},
				{domainName: "EndDate", boltName: "end", ft: stringType, leaf: false, autoInc: false},
				{domainName: "Form", boltName: "form", ft: stringType, leaf: false, autoInc: false},
			},
			newConvStruct: func() any { return &_refJsonFiling{} },
		},
		`UiSetting`: {
			rootBucket: "settings",
			buckets: []_refKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: true, autoInc: false},
			},
			newConvStruct: func() any { return &_refJsonUiSetting{} },
		},

		// End metadata
	}
)
