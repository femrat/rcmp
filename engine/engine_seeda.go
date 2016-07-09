package engine

import (
	"fmt"
	"github.com/femrat/rcmp/report"
	"io"
	"strconv"
)

type seedaEngine struct {
	basicEngine
	results [][]seedaResultCell // rowIdx, rIdx
	summary []seedaResultCell   // rIdx
}

type seedaResultCell struct {
	BestOptS string
	BestOptI int64

	AvgOptS, AvgTimeS string
	AvgOptF, AvgTimeF float64

	WinBestOpt, WinAvgOpt bool
}

func (e *seedaEngine) makeResult() error {
	// Instance-file BestOpt AvgOpt AvgTime

	e.results = make([][]seedaResultCell, len(e.rc[0].Rows))
	for rIdx := range e.results {
		e.results[rIdx] = make([]seedaResultCell, len(e.rc))
	}
	e.summary = make([]seedaResultCell, len(e.rc))

	var err error

	for rowIdx := range e.rc[0].Rows {
		for rIdx := range e.rc {
			if len(e.rc[rIdx].Rows[rowIdx].Values) != 3 {
				return fmt.Errorf("Parse line %d of report %s failed, format not match", rowIdx+1, e.rc[rIdx].ReportFile)
			}

			cur := &e.results[rowIdx][rIdx]

			// strings
			cur.BestOptS = e.rc[rIdx].Rows[rowIdx].Values[0]
			cur.AvgOptS = e.rc[rIdx].Rows[rowIdx].Values[1]
			cur.AvgTimeS = e.rc[rIdx].Rows[rowIdx].Values[2]

			// parse bestOpt
			cur.BestOptI, err = strconv.ParseInt(cur.BestOptS, 10, 64)
			if err != nil {
				return fmt.Errorf("Parse line %d of report %s failed, cannot parse BestOpt", rowIdx+1, e.rc[rIdx].ReportFile)
			}
			if !e.funcs.IsValid(cur.BestOptI) {
				return fmt.Errorf("Parse line %d of report %s failed, invalid BestOpt", rowIdx+1, e.rc[rIdx].ReportFile)
			}

			// parse avgOpt
			cur.AvgOptF, err = strconv.ParseFloat(cur.AvgOptS, 64)
			if err != nil {
				return fmt.Errorf("Parse line %d of report %s failed, cannot parse AvgOpt", rowIdx+1, e.rc[rIdx].ReportFile)
			}
			if !e.funcs.IsValidAvgFloat(cur.AvgOptF) {
				return fmt.Errorf("Parse line %d of report %s failed, invalid AvgOpt", rowIdx+1, e.rc[rIdx].ReportFile)
			}

			// parse avgTime
			cur.AvgTimeF, err = strconv.ParseFloat(cur.AvgTimeS, 64)
			if err != nil {
				return fmt.Errorf("Parse line %d of report %s failed, cannot parse AvgTime", rowIdx+1, e.rc[rIdx].ReportFile)
			}
			if cur.AvgTimeF < 0 {
				return fmt.Errorf("Parse line %d of report %s failed, AvgTime<0", rowIdx+1, e.rc[rIdx].ReportFile)
			}

			// update summary
			e.summary[rIdx].AvgOptF += cur.AvgOptF
			e.summary[rIdx].AvgTimeF += cur.AvgTimeF
		}

		markBest(len(e.rc), func(i, j int) int64 {
			return e.funcs.IsBetterAvgFloat(e.results[rowIdx][i].AvgOptF, e.results[rowIdx][j].AvgOptF)
		}, func(i int) { e.results[rowIdx][i].WinAvgOpt = true })

		markBest(len(e.rc), func(i, j int) int64 {
			return e.funcs.IsBetter(e.results[rowIdx][i].BestOptI, e.results[rowIdx][j].BestOptI)
		}, func(i int) { e.results[rowIdx][i].WinBestOpt = true })
	}

	for rIdx, rep := range e.rc {
		e.summary[rIdx].AvgOptF /= float64(len(rep.Rows))
		e.summary[rIdx].AvgTimeF /= float64(len(rep.Rows))
	}
	markBest(len(e.rc), func(i, j int) int64 {
		return e.funcs.IsBetterAvgFloat(e.summary[i].AvgOptF, e.summary[j].AvgOptF)
	}, func(i int) { e.summary[i].WinAvgOpt = true })

	return nil
}

func (e *seedaEngine) Run(writer io.Writer, rc []*report.Report) error {
	if err := e.init(rc, ""); err != nil {
		return err
	}

	if err := e.makeResult(); err != nil {
		return err
	}

	table := struct {
		ReportHeader []basicReportHeaderCell
		InstanceFile []basicInstanceFileCell
		Results      [][]seedaResultCell
		Summary      []seedaResultCell // all strings, and BestOptI are invalid
	}{
		ReportHeader: e.makeReportHeader(),
		InstanceFile: e.makeInstanceFile(nil),
		Results:      e.results,
		Summary:      e.summary,
	}

	return e.theTemplate.Execute(writer, table)
}

func (e *seedaEngine) Name() string { return "seeda" }

func (e *seedaEngine) AdditionalHelp() string {
	return "Analysis the results from different seeds"
}

func init() {
	addEngine(new(seedaEngine))
}
