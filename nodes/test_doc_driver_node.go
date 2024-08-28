package nodes

import (
	"cmp"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hackborn/onefunc/jacl"
	"github.com/hackborn/onefunc/pipeline"

	"github.com/hackborn/doc"
	"github.com/hackborn/doc_drivers/domain"
	"github.com/hackborn/doc_drivers/domain2"
	"github.com/hackborn/doc_drivers/registry"
)

func newTestDocDriverNode(docDriverPrefix string) pipeline.Node {
	n := &testDocDriverNode{}
	n.docDriverPrefix = docDriverPrefix
	return n
}

type testDocDriverNode struct {
	testDocDriverData
}

type testDocDriverData struct {
	Verbose bool
	Backend string

	docDriverPrefix string
}

func (d *testDocDriverData) docDriverName() string {
	return d.docDriverPrefix + "/" + d.Backend
}

func (n *testDocDriverNode) Start(input pipeline.StartInput) error {
	data := n.testDocDriverData
	input.SetNodeData(&data)
	return nil
}

func (n *testDocDriverNode) Run(state *pipeline.State, input pipeline.RunInput, output *pipeline.RunOutput) error {
	data := state.NodeData.(*testDocDriverData)
	var err error
	for _, pin := range input.Pins {
		switch t := pin.Payload.(type) {
		case *pipeline.ContentData:
			err = cmp.Or(err, n.runContent(data, t))
		}
	}
	return err
}

func (n *testDocDriverNode) runContent(data *testDocDriverData, cd *pipeline.ContentData) error {
	// Load tests
	entries := []testEntry{}
	err := json.Unmarshal([]byte(cd.Data), &entries)
	if err != nil {
		return err
	}

	// Open the database
	f, ok := registry.Find(data.Backend)
	if !ok {
		return fmt.Errorf("No backend named \"%v\"", data.Backend)
	}
	db, err := doc.Open(data.docDriverName(), f.DbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Run tests
	for i, te := range entries {
		err = cmp.Or(err, n.errWrap(n.runTest(db, te), cd.Name, i))
	}
	return err
}

func (n *testDocDriverNode) errWrap(err error, filename string, index int) error {
	if err != nil {
		err = fmt.Errorf("%v/%v %w", filename, index, err)
	}
	return err
}

func (n *testDocDriverNode) runTest(db *doc.DB, te testEntry) error {
	switch te.Command {
	case "get":
		return n.runGetTest(db, te)
	case "set":
		return n.runSetTest(db, te)
	case "delete":
		return n.runDeleteTest(db, te)
	default:
		return fmt.Errorf("Unhandled test command \"%v\"", te.Command)
	}
}

func (n *testDocDriverNode) runGetTest(db *doc.DB, te testEntry) error {
	switch te.Type {
	case "CollectionSetting":
		return runGetTest[domain.CollectionSetting](db, te)
	case "Events":
		return runGetTest[domain.Events](db, te)
	case "Filing":
		return runGetTest[domain.Filing](db, te)
	case "UiSetting":
		return runGetTest[domain2.UiSetting](db, te)
	default:
		return fmt.Errorf("Unhandled type \"%v\" for get", te.Type)
	}
}

func (n *testDocDriverNode) runSetTest(db *doc.DB, te testEntry) error {
	switch te.Type {
	case "CollectionSetting":
		return runSetTest[domain.CollectionSetting](db, te)
	case "Events":
		return runSetTest[domain.Events](db, te)
	case "Filing":
		return runSetTest[domain.Filing](db, te)
	case "UiSetting":
		return runSetTest[domain2.UiSetting](db, te)
	default:
		return fmt.Errorf("Unhandled type \"%v\" for set", te.Type)
	}
}

func (n *testDocDriverNode) runDeleteTest(db *doc.DB, te testEntry) error {
	switch te.Type {
	case "CollectionSetting":
		return runDeleteTest[domain.CollectionSetting](db, te)
	case "Events":
		return runDeleteTest[domain.Events](db, te)
	case "Filing":
		return runDeleteTest[domain.Filing](db, te)
	case "UiSetting":
		return runDeleteTest[domain2.UiSetting](db, te)
	default:
		return fmt.Errorf("Unhandled type \"%v\" for delete", te.Type)
	}
}

func runGetTest[T any](db *doc.DB, te testEntry) error {
	req := doc.GetRequest{}
	var err error
	req.Condition, err = db.Expr(te.Expr, nil).Compile()
	if err != nil {
		return err
	}
	resp, err := doc.Get[T](db, req)
	if err != nil {
		return err
	}
	/*
		fmt.Println("got")
		for _, item := range resp.Results {
			fmt.Println("\t", item)
		}
	*/
	return jacl.Run(resp.Results, te.Response...)
}

func runSetTest[T any](db *doc.DB, te testEntry) error {
	fitem, err := newTestItem[T](te.Item)
	if err != nil {
		return err
	}
	req := doc.SetRequest[T]{Item: fitem, Filter: te.MakeFilter()}
	resp, err := doc.Set(db, req)
	// The API is currently unclear on whether a return item
	// is required, but I think all the drivers ignore it right
	// now so we'll just assume it's optional.
	if resp.Item == nil {
		return nil
	}
	return jacl.Run(resp.Item, te.Response...)
}

func runDeleteTest[T any](db *doc.DB, te testEntry) error {
	item, err := newTestItem[T](te.Item)
	if err != nil {
		return err
	}
	req := doc.DeleteRequest[T]{Item: item}
	_, err = doc.Delete[T](db, req)
	return err
}

func newTestItem[T any](item map[string]any) (T, error) {
	var t T
	b, err := json.Marshal(item)
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(b, &t)
	//	fmt.Println(t, item)
	return t, err
}

type testEntry struct {
	Command  string         `json:"command"`
	Type     string         `json:"type"`
	Expr     string         `json:"expr"`
	Item     map[string]any `json:"item"`
	Filter   string         `json:"filter"`
	Response []string       `json:"response"`
}

func (e testEntry) MakeFilter() doc.Filter {
	f := strings.ToLower(e.Filter)
	if f == "createitem" {
		return doc.FilterCreateItem
	}
	return doc.Filter{}
}
