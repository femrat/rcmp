package engine

import (
	"fmt"
	"github.com/femrat/rcmp/report"
	"io"
	"strconv"
)

type seedEngine struct {
	basicEngine

	results []seedResultCell
	avgOpt  float64
	avgTime float64
}

type seedResultCell struct {
	AvgOpt  float64
	AvgTime float64
	BestOpt int64
}

func (e *seedEngine) makeResult() error {
	e.results = make([]seedResultCell, len(e.rc[0].Rows))
	for rIdx, rep := range e.rc {
		for rowIdx, row := range rep.Rows {
			// prase
			opt, err := strconv.ParseInt(row.Values[0], 10, 64)
			if err != nil {
				return fmt.Errorf("Parse line %d of report %s failed", rowIdx+1, rep.ReportFile)
			}
			time, err := strconv.ParseFloat(row.Values[1], 64)
			if err != nil {
				return fmt.Errorf("Parse line %d of report %s failed", rowIdx+1, rep.ReportFile)
			}

			// verify
			if !e.funcs.IsValid(opt) {
				return fmt.Errorf("The opt value in line %d of report %s is invalid", rowIdx+1, rep.ReportFile)
			}

			e.avgOpt += float64(opt)
			e.avgTime += float64(time)

			curResult := &e.results[rowIdx]
			if rIdx == 0 || e.funcs.IsBetter(opt, curResult.BestOpt) > 0 {
				curResult.BestOpt = opt
			}
			curResult.AvgOpt += float64(opt)
			curResult.AvgTime += float64(time)
		}
	}

	for i := range e.results {
		e.results[i].AvgOpt /= float64(len(e.rc))
		e.results[i].AvgTime /= float64(len(e.rc))
	}

	e.avgOpt /= float64(len(e.rc) * len(e.rc[0].Rows))
	e.avgTime /= float64(len(e.rc) * len(e.rc[0].Rows))

	return nil
}

func (e *seedEngine) Run(writer io.Writer, rc []*report.Report) error {
	if err := e.init(rc, ""); err != nil {
		return err
	}

	if err := e.makeResult(); err != nil {
		return err
	}

	table := struct {
		ReportHeader []basicReportHeaderCell
		InstanceFile []basicInstanceFileCell
		Results      []seedResultCell
		AvgOpt       float64
		AvgTime      float64
	}{
		ReportHeader: e.makeReportHeader(),
		InstanceFile: e.makeInstanceFile(nil),
		Results:      e.results,
		AvgOpt:       e.avgOpt,
		AvgTime:      e.avgTime,
	}

	return e.theTemplate.Execute(writer, table)
}

func (e *seedEngine) Name() string { return "seed" }

func (e *seedEngine) AdditionalHelp() string {
	return "Collect the results running with different seeds"
}

func init() {
	addEngine(new(seedEngine))
}
