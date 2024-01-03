package domain

// struct Filing represents a single filing for a company.
type Filing struct {
	Ticker string `tox:"ticker, key:a"`
	// End date of the filing period.
	EndDate string `json:"end" tox:"end, key:b"`
	// Form used in the filing
	Form string `json:"form" tox:"form, key:c"`
	// Amount of filing.
	Value int64 `json:"val" tox:"val"`
	// Units used for the value (i.e. "usd").
	Units string `json:"units" tox:"units"`
	// Fiscal year of the filing
	FiscalYear int `json:"fy" tox:"fy"`
}
