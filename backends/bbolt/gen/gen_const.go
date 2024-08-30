package bboltgendriver

// autogenerated with github.com/hackborn/doc_drivers
// do not modify

type boltKey = []byte

var genKeySep = []byte("/")

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
	genMetadatas = map[string]*genMetadata{

		`CollectionSetting`: {
			rootBucket: "settings",
			buckets: []genKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: true, flags: 0},
			},
			newConvStruct: func() any { return &genJsonCollectionSetting{} },
		},
		`Company`: {
			rootBucket: "company",
			buckets: []genKeyMetadata{
				{domainName: "Id", boltName: "id", ft: stringType, leaf: false, flags: 0},
				{domainName: "Name", boltName: "name", ft: stringType, leaf: false, flags: 0},
				{domainName: "FoundedYear", boltName: "fy", ft: stringType, leaf: false, flags: 0},
			},
			newConvStruct: func() any { return &genJsonCompany{} },
		},
		`Events`: {
			rootBucket: "events",
			buckets: []genKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: false, flags: 0},
				{domainName: "Time", boltName: "time", ft: uint64Type, leaf: true, flags: 1},
			},
			newConvStruct: func() any { return &genJsonEvents{} },
		},
		`FavouritesSetting`: {
			rootBucket: "settings",
			buckets: []genKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: true, flags: 0},
			},
			newConvStruct: func() any { return &genJsonFavouritesSetting{} },
		},
		`Filing`: {
			rootBucket: "filing",
			buckets: []genKeyMetadata{
				{domainName: "Ticker", boltName: "ticker", ft: stringType, leaf: false, flags: 0},
				{domainName: "EndDate", boltName: "end", ft: stringType, leaf: false, flags: 0},
				{domainName: "Form", boltName: "form", ft: stringType, leaf: false, flags: 0},
			},
			newConvStruct: func() any { return &genJsonFiling{} },
		},
		`UiSetting`: {
			rootBucket: "settings",
			buckets: []genKeyMetadata{
				{domainName: "Name", boltName: "name", ft: stringType, leaf: true, flags: 0},
			},
			newConvStruct: func() any { return &genJsonUiSetting{} },
		},
	}
)
