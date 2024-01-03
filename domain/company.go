package domain

// Company stores standard company data.
type Company struct {
	// A unique ID for the company.
	Id string `tox:"id, key"`
	// Friendly name. Designed as a secondary index.
	Name string `tox:"name, key:b"`
	// Value of the company (in some unknown units).
	Value int64 `json:"val" tox:"val"`
	// Year the company was founded.
	FoundedYear int `json:"fy" tox:"fy"`
}
