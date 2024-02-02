package nodes

import (
	"encoding/json"
	"fmt"

	"github.com/hackborn/doc"
	"github.com/hackborn/doc_drivers/domain"
	"github.com/hackborn/doc_drivers/registry"
	"github.com/hackborn/onefunc/pipeline"
)

func newRunDocDriverNode(docDriverPrefix string) pipeline.Node {
	n := &runDocDriverNode{docDriverPrefix: docDriverPrefix}
	n.fn = n.makeReports()
	return n
}

type runDocDriverNode struct {
	Verbose bool
	Backend string

	docDriverPrefix string
	fn              []runReportFunc
}

func (n *runDocDriverNode) Run(state *pipeline.State, input pipeline.RunInput) (*pipeline.RunOutput, error) {
	if state.Flush {
		return nil, nil
	}
	f, ok := registry.Find(n.Backend)
	if !ok {
		return nil, fmt.Errorf("No backend named \"%v\"", n.Backend)
	}

	// Open the database
	db, err := doc.Open(n.docDriverName(), f.DbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return n.runReport(db)
}

func (n *runDocDriverNode) runReport(db *doc.DB) (*pipeline.RunOutput, error) {
	output := &pipeline.RunOutput{}
	report := &RunReportData{}
	for _, fn := range n.fn {
		e := fn(db)
		if n.Verbose {
			sr := ""
			if e.Response != nil {
				resp, _ := json.Marshal(e.Response)
				sr = string(resp)
			}
			fmt.Println(e.Name, "resp", sr, "err", e.Err)
		}
		if e.Err != nil {
			return nil, fmt.Errorf("Error for driver \"%v\" test \"%v\": %w", n.docDriverName(), e.Name, e.Err)
		}
		report.Entries = append(report.Entries, e)
	}

	output.Pins = append(output.Pins, pipeline.Pin{Name: n.docDriverName(), Payload: report})
	return output, nil
}

func (n *runDocDriverNode) makeReports() []runReportFunc {
	fn := []runReportFunc{
		func(db *doc.DB) ReportEntry {
			filing := domain.Filing{Ticker: "AAPL", EndDate: "2023", Form: "wd-40", Value: 10000, Units: "usd"}
			req := doc.SetRequest[domain.Filing]{Item: filing}
			resp, err := doc.Set(db, req)
			return ReportEntry{Name: "Set Filing 1", Response: resp, Err: err}
		},
		func(db *doc.DB) ReportEntry {
			filing := domain.Filing{Ticker: "GOOG", EndDate: "2023", Form: "wd-40", Value: 10000, Units: "usd"}
			req := doc.SetRequest[domain.Filing]{Item: filing}
			resp, err := doc.Set(db, req)
			return ReportEntry{Name: "Set Filing 2", Response: resp, Err: err}
		},
		func(db *doc.DB) ReportEntry {
			filing := domain.Filing{Ticker: "GOOG", EndDate: "2022", Form: "wd-40", Value: 10000, Units: "usd"}
			req := doc.SetRequest[domain.Filing]{Item: filing}
			resp, err := doc.Set(db, req)
			return ReportEntry{Name: "Set Filing 3", Response: resp, Err: err}
		},
		func(db *doc.DB) ReportEntry {
			filing := domain.Filing{Ticker: "GOOG", EndDate: "2022", Form: "wd-40", Value: 10010, Units: "usd"}
			req := doc.SetRequest[domain.Filing]{Item: filing}
			resp, err := doc.Set(db, req)
			return ReportEntry{Name: "Set Filing 4", Response: resp, Err: err}
		},
		func(db *doc.DB) ReportEntry {
			getreq := doc.GetRequest{}
			getreq.Condition, _ = db.Expr(`ticker = "GOOG" AND form = "wd-40"`, nil).Compile()
			resp, err := doc.Get[domain.Filing](db, getreq)
			return ReportEntry{Name: "Get Filing", Response: resp, Err: err}
		},
		func(db *doc.DB) ReportEntry {
			getonereq := doc.GetRequest{}
			getonereq.Condition, _ = db.Expr(`ticker = GOOG AND end = 2022 AND form = "wd-40"`, nil).Compile()
			resp, err := doc.GetOne[domain.Filing](db, getonereq)
			return ReportEntry{Name: "GetOne Filing 1", Response: resp, Err: err}
		},
		func(db *doc.DB) ReportEntry {
			filing := domain.Filing{Ticker: "GOOG", EndDate: "2022", Form: "wd-40", Value: 10010, Units: "usd"}
			delreq := doc.DeleteRequest[domain.Filing]{Item: filing}
			resp, err := doc.Delete[domain.Filing](db, delreq)
			return ReportEntry{Name: "Delete Filing 1", Response: resp, Err: err}
		},
		func(db *doc.DB) ReportEntry {
			getonereq := doc.GetRequest{}
			getonereq.Condition, _ = db.Expr(`ticker = GOOG AND end = 2022 AND form = "wd-40"`, nil).Compile()
			resp, err := doc.GetOne[domain.Filing](db, getonereq)
			return ReportEntry{Name: "GetOne Filing 2", Response: resp, Err: err}
		},
	}

	return fn
}

func (n *runDocDriverNode) docDriverName() string {
	return n.docDriverPrefix + "/" + n.Backend
}

type runReportFunc func(db *doc.DB) ReportEntry
