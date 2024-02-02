package nodes

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hackborn/onefunc/pipeline"
)

type compareReportsNode struct {
}

func (n *compareReportsNode) Run(state *pipeline.State, input pipeline.RunInput) (*pipeline.RunOutput, error) {
	if state.Flush == true {
		return nil, nil
	}
	var refRun *RunReportData = nil
	var genRun *RunReportData = nil
	for _, pin := range input.Pins {
		switch p := pin.Payload.(type) {
		case *RunReportData:
			if strings.HasPrefix(pin.Name, "ref/") {
				refRun = p
			} else if strings.HasPrefix(pin.Name, "gen/") {
				genRun = p
			}
		}
	}
	if refRun == nil {
		return nil, fmt.Errorf("Missing input pin ref/*")
	}
	if genRun == nil {
		return nil, fmt.Errorf("Missing input pin gen/*")
	}
	return n.compare(refRun, genRun)
}

func (n *compareReportsNode) compare(refRun, genRun *RunReportData) (*pipeline.RunOutput, error) {
	if len(refRun.Entries) < 1 || len(refRun.Entries) != len(genRun.Entries) {
		return nil, fmt.Errorf("Missing report entries")
	}
	output := &pipeline.RunOutput{}
	for i, re := range refRun.Entries {
		ge := genRun.Entries[i]
		err := n.compareEntries(re, ge)
		if err != nil {
			return nil, err
		}
	}
	fmt.Println("ok")
	return output, nil
}

func (n *compareReportsNode) compareEntries(refEntry, genEntry ReportEntry) error {
	if refEntry.Name != genEntry.Name {
		return fmt.Errorf("Wrong test name, have %v but want %v", genEntry.Name, refEntry.Name)
	}
	if refEntry.Err == nil && genEntry.Err != nil {
		return fmt.Errorf("%v wrong error, have %w but want nil", refEntry.Name, genEntry.Err)
	} else if refEntry.Err != nil && genEntry.Err == nil {
		return fmt.Errorf("%v wrong error, have nil but want %w", refEntry.Name, refEntry.Err)
	}
	if !reflect.DeepEqual(refEntry.Response, genEntry.Response) {
		d1, _ := json.Marshal(refEntry.Response)
		d2, _ := json.Marshal(genEntry.Response)
		return fmt.Errorf("%v wrong response, have %v but want %v", refEntry.Name, string(d2), string(d1))
	}
	return nil
}
