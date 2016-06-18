package engine

import (
	"flag"
	"fmt"
	"github.com/femrat/rcmp/report"
	"io"
	"strconv"
)

type sEngine struct {
	basicEngine
	sRenameFile string
}

// ==============
// Output Structs
// ==============

type sResultCell struct {
	Type string // Result or BaseResult

	IsValid bool
	opt     int64
	rawTime string
	time    float64

	IsBest bool // if its opt is the best of the line

	// for non-base only
	Base          *sResultCell // base one is pointed to nil
	compareToBase string       // better, worse, equal, bothNA
}

func (s *sResultCell) valid() {
	if !s.IsValid {
		panic("must be valid to get values")
	}
}
func (s *sResultCell) Opt() int64 {
	s.valid()
	return s.opt
}
func (s *sResultCell) RawTime() string {
	s.valid()
	return s.rawTime
}
func (s *sResultCell) Time() float64 {
	s.valid()
	return s.time
}
func (s *sResultCell) CompareToBase() string {
	if s.Type != "Result" {
		panic("only type==result can get IsBaseValid()")
	}
	return s.compareToBase
}
func (s *sResultCell) OptDiff() int64 {
	if s.Type != "Result" {
		panic("only type==result can calculate OptDiff")
	}
	return s.Opt() - s.Base.Opt() // will check if both are valid
}

// ====================
// Unexported Functions
// ====================

func (e *sEngine) parseOneReportRow(rowIdx int) ([]sResultCell, error) {
	// fill Type and sValues field
	rowData := make([]sResultCell, len(e.rc))

	// parse
	for rIdx, rep := range e.rc {
		if rIdx == 0 {
			rowData[rIdx].Type = "BaseResult"
		} else {
			rowData[rIdx].Type = "Result"
		}
		if len(rep.Rows[rowIdx].Values) != 2 {
			return nil, fmt.Errorf("The line %d of report %s does not fit 'instance-file opt opt-time' format", rowIdx, rep.ReportFile)
		}
		var err error

		if rowData[rIdx].opt, err = strconv.ParseInt(rep.Rows[rowIdx].Values[0], 10, 64); err != nil {
			return nil, err
		}
		rowData[rIdx].rawTime = rep.Rows[rowIdx].Values[1]
		if rowData[rIdx].time, err = strconv.ParseFloat(rep.Rows[rowIdx].Values[1], 64); err != nil {
			return nil, err
		}

		rowData[rIdx].IsValid = e.funcs.IsValid(rowData[rIdx].opt)
	}

	return rowData, nil
}

func (e *sEngine) makeOneReportRow(rowIdx int, counting []basicCountingCell) ([]sResultCell, error) {
	row, err := e.parseOneReportRow(rowIdx)
	if err != nil {
		return nil, err
	}

	// find best opt, count NA
	var foundBest bool = false
	var bestOpt int64
	for rIdx := range e.rc {
		if row[rIdx].IsValid {
			if !foundBest || e.funcs.IsBetter(row[rIdx].opt, bestOpt) > 0 {
				bestOpt = row[rIdx].opt
				foundBest = true
			}
		} else {
			counting[rIdx].increase("NA")
		}
	}

	for rIdx := range e.rc {
		row[rIdx].IsBest = foundBest && row[rIdx].opt == bestOpt
		if rIdx > 0 { // skip for base
			row[rIdx].Base = &row[0]

			if row[0].IsValid {
				if row[rIdx].IsValid {
					res := e.funcs.IsBetter(row[rIdx].opt, row[0].opt)
					if res > 0 {
						row[rIdx].compareToBase = "better"
						counting[rIdx].increase("Better")
					} else if res == 0 {
						row[rIdx].compareToBase = "equal"
						counting[rIdx].increase("Equal")
						counting[rIdx].increase("EqualWithNA")
					} else {
						row[rIdx].compareToBase = "Worse"
						counting[rIdx].increase("Worse")
					}
				} else { // row[0].IsValid && !row[rIdx].IsValid
					row[rIdx].compareToBase = "worse"
					counting[rIdx].increase("Worse")
				}
			} else { // !row[0].IsValid
				if row[rIdx].IsValid {
					row[rIdx].compareToBase = "better"
					counting[rIdx].increase("Better")
				} else {
					row[rIdx].compareToBase = "bothNA"
					counting[rIdx].increase("BothNA")
					counting[rIdx].increase("EqualWithNA")
				}
			}
		}
	}

	// Delta is not calculated here, which will be done in makeReportsAndCounting()

	return row, nil
}

func (e *sEngine) makeReportsAndCounting() ([][]sResultCell, []basicCountingCell, error) {
	table := make([][]sResultCell, len(e.rc[0].Rows)) // [row][col]. [0] is base

	counting := make([]basicCountingCell, len(e.rc)) // [0] is base, limited element is available.
	for i := range counting {
		counting[i].init()
	}

	for rowIdx := range e.rc[0].Rows {
		if row, err := e.makeOneReportRow(rowIdx, counting); err != nil {
			return nil, nil, err
		} else {
			table[rowIdx] = row
		}
	}

	basicCountingCellUpdates(counting)

	return table, counting, nil
}

func (e *sEngine) init(rc []*report.Report) error {
	e.rc = rc

	if err := e.loadTemplate(""); err != nil {
		return err
	}
	if err := e.basicEngine.parseCompareFunc(); err != nil {
		return err
	}

	return nil
}

// ==================
// Exported Functions
// ==================

func (e *sEngine) Run(writer io.Writer, rc []*report.Report) error {
	if err := e.init(rc); err != nil {
		return err
	}

	table := struct {
		ReportHeader []basicReportHeaderCell
		InstanceFile []basicInstanceFileCell
		Results      [][]sResultCell // [row index][report index]. the first report is Base
		Counting     []basicCountingCell
	}{
		ReportHeader: e.makeReportHeader(),
	}
	var err error

	if e.sRenameFile != "" {
		var rename Replacer
		if err = rename.LoadFromDisk(e.sRenameFile); err != nil {
			return err
		}
		table.InstanceFile = e.makeInstanceFile(func(old string) string { return rename.Replace(old) })
	} else {
		table.InstanceFile = e.makeInstanceFile(nil)
	}

	table.Results, table.Counting, err = e.makeReportsAndCounting()
	if err != nil {
		return err
	}

	return e.theTemplate.Execute(writer, table)
}

func (e *sEngine) Name() string { return "s" }

func (e *sEngine) AdditionalHelp() string {
	return "Compare separately. The first report is the base, and others compare to the base separately."
}

func (e *sEngine) SetFlags(flag *flag.FlagSet) {
	e.basicEngine.SetFlags(flag)
	flag.StringVar(&e.sRenameFile, "rename", "", "Replace the instance-file with new name following the `FILE`.")
}

func init() {
	addEngine(new(sEngine))
}
