package bboltrefdriver

// Begin json

type _refJsonCollectionSetting struct {
	Value []int64 `json:"value"`
}

type _refJsonCompany struct {
	Value int64 `json:"val"`
}

type _refJsonEvents struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type _refJsonFiling struct {
	Value      int64  `json:"val"`
	Units      string `json:"units"`
	FiscalYear int    `json:"fy"`
}

type _refJsonUiSetting struct {
	Value map[string]string `json:"value"`
}

// End json
