package domain

// struct Filing represents a single filing for a company.
type Filing struct {
	Ticker string `doc:"ticker, key:a"`
	// End date of the filing period.
	EndDate string `json:"end" doc:"end, key:b"`
	// Form used in the filing
	Form string `json:"form" doc:"form, key:c"`
	// Amount of filing.
	Value int64 `json:"val" doc:"val"`
	// Units used for the value (i.e. "usd").
	Units string `json:"units" doc:"units"`
	// Fiscal year of the filing
	FiscalYear int `json:"fy" doc:"fy"`
}
