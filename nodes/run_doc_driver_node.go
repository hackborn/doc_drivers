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
	n := &runDocDriverNode{}
	n.docDriverPrefix = docDriverPrefix
	n.fn = n.makeReports()
	return n
}

type runDocDriverNode struct {
	runDocDriverData

	fn []runReportFunc
}

type runDocDriverData struct {
	Verbose bool
	Backend string

	docDriverPrefix string
}

func (d *runDocDriverData) docDriverName() string {
	return d.docDriverPrefix + "/" + d.Backend
}

func (n *runDocDriverNode) Start(input pipeline.StartInput) error {
	data := n.runDocDriverData
	input.SetNodeData(&data)
	return nil
}

func (n *runDocDriverNode) Run(state *pipeline.State, input pipeline.RunInput, output *pipeline.RunOutput) error {
	data := state.NodeData.(*runDocDriverData)
	f, ok := registry.Find(data.Backend)
	if !ok {
		return fmt.Errorf("No backend named \"%v\"", data.Backend)
	}

	// Open the database
	db, err := doc.Open(data.docDriverName(), f.DbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	return n.runReport(db, data, output)
}

func (n *runDocDriverNode) runReport(db *doc.DB, data *runDocDriverData, output *pipeline.RunOutput) error {
	report := &RunReportData{}
	for _, fn := range n.fn {
		e := fn(db)
		if data.Verbose {
			sr := ""
			if e.Response != nil {
				resp, _ := json.Marshal(e.Response)
				sr = string(resp)
			}
			fmt.Println(e.Name, "resp", sr, "err", e.Err)
		}
		if e.Err != nil {
			return fmt.Errorf("Error for driver \"%v\" test \"%v\": %w", data.docDriverName(), e.Name, e.Err)
		}
		report.Entries = append(report.Entries, e)
	}

	output.Pins = append(output.Pins, pipeline.Pin{Name: data.docDriverName(), Payload: report})
	return nil
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
		func(db *doc.DB) ReportEntry {
			// test serialized writing
			value := []int64{4, 7, 8}
			setting := domain.CollectionSetting{Name: "favs", Value: value}
			req := doc.SetRequest[domain.CollectionSetting]{Item: setting}
			resp, err := doc.Set(db, req)
			return ReportEntry{Name: "Set Setting 1", Response: resp, Err: err}
		},
		func(db *doc.DB) ReportEntry {
			// test serialized reading
			getreq := doc.GetRequest{}
			getreq.Condition, _ = db.Expr(`name = "favs"`, nil).Compile()
			resp, err := doc.Get[domain.CollectionSetting](db, getreq)
			return ReportEntry{Name: "Get Fav Setting", Response: resp, Err: err}
		},
	}

	return fn
}

type runReportFunc func(db *doc.DB) ReportEntry
