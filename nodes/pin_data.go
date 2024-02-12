package nodes

import (
	"github.com/hackborn/onefunc/pipeline"
)

type RunReportData struct {
	Entries []ReportEntry
}

func (d *RunReportData) Clone() pipeline.Cloner {
	dst := *d
	return &dst
}

type ReportEntry struct {
	Name     string
	Response any
	Err      error
}
