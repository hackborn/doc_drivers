package nodes

type RunReportData struct {
	Entries []ReportEntry
}

type ReportEntry struct {
	Name     string
	Response any
	Err      error
}
