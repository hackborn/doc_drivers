package domain

// Company stores standard company data.
type Company struct {
	// A unique ID for the company.
	Id string `doc:"id, key"`
	// Friendly name. Designed as a secondary index.
	Name string `doc:"name, key(b)"`
	// Value of the company (in some unknown units).
	Value int64 `json:"val" doc:"val"`
	// Year the company was founded.
	FoundedYear int `json:"fy" doc:"fy, key(c,1)"`
	// Do not include this in the driver.
	Skip int `doc:"-"`
}
