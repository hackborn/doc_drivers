package main

import (
	"fmt"

	"github.com/hackborn/doc"
	"github.com/hackborn/doc_drivers"
	"github.com/hackborn/doc_drivers/domain"
	"github.com/hackborn/onefunc/errors"
)

// main loads and runs a driver, performing several tests on it.
func main() {
	f, err := drivers.GetFactoryFromCla()
	if err != nil {
		fmt.Println(err)
		fmt.Println(help)
		return
	}
	fmt.Println(f)

	// Initialize the driver
	doc.Register(f.DriverName, f.New())

	// Open the database
	db, err := doc.Open(f.DriverName, f.DbPath)
	errors.Panic(err)
	defer db.Close()

	// Run the tests
	runTests(db)
}

func runTests(db *doc.DB) {
	filing := domain.Filing{Ticker: "AAPL", EndDate: "2023", Form: "wd-40", Value: 10000, Units: "usd"}
	req := doc.SetRequest[domain.Filing]{Item: filing}
	_, err := doc.Set(db, req)
	fmt.Println("Set Filing 1 err", err)

	filing = domain.Filing{Ticker: "GOOG", EndDate: "2023", Form: "wd-40", Value: 10000, Units: "usd"}
	req = doc.SetRequest[domain.Filing]{Item: filing}
	_, err = doc.Set(db, req)
	fmt.Println("Set Filing 2 err", err)

	filing = domain.Filing{Ticker: "GOOG", EndDate: "2022", Form: "wd-40", Value: 10000, Units: "usd"}
	req = doc.SetRequest[domain.Filing]{Item: filing}
	_, err = doc.Set(db, req)
	fmt.Println("Set Filing 3 err", err)

	filing = domain.Filing{Ticker: "GOOG", EndDate: "2022", Form: "wd-40", Value: 10010, Units: "usd"}
	req = doc.SetRequest[domain.Filing]{Item: filing}
	_, err = doc.Set(db, req)
	fmt.Println("Set Filing 4 err", err)

	getreq := doc.GetRequest{}
	getreq.Condition, _ = db.Expr(`ticker = "GOOG" AND form = "wd-40"`, nil).Compile()
	resp2, err := doc.Get[domain.Filing](db, getreq)
	fmt.Println("Get Filing resp", resp2, "err", err)
	for _, f := range resp2.Results {
		fmt.Println("\t", *f)
	}

	getonereq := doc.GetRequest{}
	getonereq.Condition, _ = db.Expr(`ticker = GOOG AND end = 2022 AND form = "wd-40"`, nil).Compile()
	oneResp, err := doc.GetOne[domain.Filing](db, getonereq)
	fmt.Println("GetOne Filing resp", oneResp, "err", err)
	if oneResp.Result != nil {
		fmt.Println("\tgetOne result", *oneResp.Result)
	}

	delreq := doc.DeleteRequest[domain.Filing]{Item: filing}
	delresp, err := doc.Delete[domain.Filing](db, delreq)
	fmt.Println("Delete Filing resp", delresp, "err", err)

	oneResp, err = doc.GetOne[domain.Filing](db, getonereq)
	fmt.Println("GetOne Filing resp", oneResp, "err", err)
	if oneResp.Result != nil {
		fmt.Println("\tgetOne result", *oneResp.Result)
	}
}
