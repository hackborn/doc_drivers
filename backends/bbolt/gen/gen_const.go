package bboltgendriver

// autogenerated with github.com/hackborn/doc_drivers
// do not modify

type fieldType uint8

const (
	stringType fieldType = iota
	uint64Type
)

var (
	genMetadatas = map[string]*genMetadata{

		`CollectionSetting`: {
			rootBucket: "settings",
			buckets: []genKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: true, autoInc: false},
			},
			newConvStruct: func() any { return &genJsonCollectionSetting{} },
		},
		`Company`: {
			rootBucket: "company",
			buckets: []genKeyMetadata{
				{domainName: "Id", boltName: "id", ft: stringType, leaf: false, autoInc: false},
				{domainName: "Name", boltName: "name", ft: stringType, leaf: false, autoInc: false},
				{domainName: "FoundedYear", boltName: "fy", ft: stringType, leaf: false, autoInc: false},
			},
			newConvStruct: func() any { return &genJsonCompany{} },
		},
		`Events`: {
			rootBucket: "events",
			buckets: []genKeyMetadata{
				{domainName: "Time", boltName: "time", ft: uint64Type, leaf: true, autoInc: true},
			},
			newConvStruct: func() any { return &genJsonEvents{} },
		},
		`Filing`: {
			rootBucket: "filing",
			buckets: []genKeyMetadata{
				{domainName: "Ticker", boltName: "ticker", ft: stringType, leaf: false, autoInc: false},
				{domainName: "EndDate", boltName: "end", ft: stringType, leaf: false, autoInc: false},
				{domainName: "Form", boltName: "form", ft: stringType, leaf: false, autoInc: false},
			},
			newConvStruct: func() any { return &genJsonFiling{} },
		},
		`UiSetting`: {
			rootBucket: "settings",
			buckets: []genKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: true, autoInc: false},
			},
			newConvStruct: func() any { return &genJsonUiSetting{} },
		},
	}
)
