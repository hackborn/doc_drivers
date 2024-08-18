package bboltrefdriver

var (
	_refMetadatas = map[string]*_refMetadata{
		// Begin metadata
		`CollectionSetting`: {
			rootBucket: "settings",
			buckets: []_refKeyMetadata{
				{domainName: "Name", boltName: "name"},
			},
			newConvStruct: func() any { return &_refJsonCollectionSetting{} },
		},
		`Company`: {
			rootBucket: "company",
			buckets: []_refKeyMetadata{
				{domainName: "Id", boltName: "id"},
				{domainName: "Name", boltName: "name"},
				{domainName: "FoundedYear", boltName: "fy"},
			},
			newConvStruct: func() any { return &_refJsonCompany{} },
		},
		`Events`: {
			rootBucket: "events",
			buckets: []_refKeyMetadata{
				{domainName: "Time", boltName: "time"},
			},
			newConvStruct: func() any { return &_refJsonEvents{} },
		},
		`Filing`: {
			rootBucket: "filing",
			buckets: []_refKeyMetadata{
				{domainName: "Ticker", boltName: "ticker"},
				{domainName: "EndDate", boltName: "end"},
				{domainName: "Form", boltName: "form"},
			},
			newConvStruct: func() any { return &_refJsonFiling{} },
		},
		`UiSetting`: {
			rootBucket: "settings",
			buckets: []_refKeyMetadata{
				{domainName: "Name", boltName: "name"},
			},
			newConvStruct: func() any { return &_refJsonUiSetting{} },
		},

		// End metadata
	}
)
