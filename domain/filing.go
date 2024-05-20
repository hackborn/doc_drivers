package domain

// struct Filing represents a single filing for a company.
type Filing struct {
	Ticker string `doc:"key(,0)"`
	// End date of the filing period.
	EndDate string `json:"end" doc:"name(end), key(,1)"`
	// Form used in the filing
	Form string `json:"form" doc:"key (,2)"`
	// Amount of filing.
	Value int64 `json:"val" doc:"name(val)"`
	// Units used for the value (i.e. "usd").
	Units string `json:"units"`
	// Fiscal year of the filing
	FiscalYear int `json:"fy" doc:"name(fy)"`
	// Private fields are treated as table specs
	_table int `doc:"name(filing)"`
}
