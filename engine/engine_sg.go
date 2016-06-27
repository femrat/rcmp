package engine

import (
	"flag"
	"fmt"
	"github.com/femrat/rcmp/report"
	"io"
	"os"
	"strconv"
)

type sgEngine struct {
	basicEngine
	groupFile string            // the rule of grouping
	groupTo   map[string]string // instance-file --> group name
	groupName []string          // the order of grouping is the order of group appears in groupFile firstly
	groupIdx  map[string]int    // group name --> the index in groupName
	groupCap  []int
}

func (e *sgEngine) init(rc []*report.Report) error {
	e.rc = rc

	if err := e.loadTemplate(""); err != nil {
		return err
	}
	if err := e.basicEngine.parseCompareFunc(); err != nil {
		return err
	}

	e.groupTo = make(map[string]string)
	e.groupIdx = make(map[string]int)

	// load group file
	if e.groupFile == "" {
		return fmt.Errorf("The groupFile (-group option) must be set")
	}
	fp, err := os.Open(e.groupFile)
	if err != nil {
		return fmt.Errorf("can't open groupFile: %v", err)
	}
	defer fp.Close()
	var inst, grp string
	for {
		n, err := fmt.Fscan(fp, &grp, &inst)
		if n == 2 {
			if idx, ok := e.groupIdx[grp]; !ok { // new group
				e.groupIdx[grp] = len(e.groupName)
				e.groupName = append(e.groupName, grp)
				e.groupCap = append(e.groupCap, 1)
			} else {
				e.groupCap[idx]++
			}
			if _, ok := e.groupTo[inst]; ok {
				return fmt.Errorf("The instance-file [%s] is double grouped", inst)
			}
			e.groupTo[inst] = grp
		} else if n == 0 {
			if err == io.EOF {
				break
			} else {
				return err
			}
		} else {
			return fmt.Errorf("GroupFile should be in pair, the group [%s] needs one more instance-file.", grp)
		}
	}

	if len(e.groupTo) != len(rc[0].Rows) {
		return fmt.Errorf("The number of instance-file and the number of pairs in groupFile does not match.")
	}

	return nil
}

func (e *sgEngine) procOneRow(rowIdx int, pCounting []basicCountingCell) error {
	parseOpt := func(rIdx int) (int64, error) {
		return strconv.ParseInt(e.rc[rIdx].Rows[rowIdx].Values[0], 10, 64)
	}

	var err error
	var baseOpt, opt int64
	var baseValid, curValid bool

	for rIdx, rep := range e.rc {
		if len(rep.Rows[rowIdx].Values) != 2 {
			return fmt.Errorf("The line %d of report %s does not fit 'instance-file opt opt-time' format", rowIdx, rep.ReportFile)
		}

		opt, err = parseOpt(rIdx)
		if err != nil {
			return fmt.Errorf("Failed to parse opt value in line %d of report %s", rowIdx, rep.ReportFile)
		}
		curValid = e.funcs.IsValid(opt)

		if !curValid {
			pCounting[rIdx].increase("NA")
		}

		if rIdx == 0 {
			baseOpt = opt
			baseValid = curValid
			continue
		}

		p := pCounting[rIdx]

		if baseValid {
			if curValid {
				res := e.funcs.IsBetter(opt, baseOpt)
				if res > 0 {
					p.increase("Better")
				} else if res == 0 {
					p.increase("Equal")
					p.increase("EqualWithNA")
				} else {
					p.increase("Worse")
				}
			} else { // baseValid && !curValid
				p.increase("Worse")
			}
		} else { // !baseValid
			if curValid {
				p.increase("Better")
			} else {
				p.increase("BothNA")
				p.increase("EqualWithNA")
			}
		}
	}
	return nil
}

func (e *sgEngine) makeCounting() (pCounting [][]basicCountingCell, fCounting []basicCountingCell, err error) {
	init := func(c []basicCountingCell) {
		for i := range c {
			c[i].init()
		}
	}

	pCounting = make([][]basicCountingCell, len(e.groupName))
	for i := range e.groupName {
		pCounting[i] = make([]basicCountingCell, len(e.rc))
		init(pCounting[i])
	}

	for rowIdx := 0; rowIdx < len(e.rc[0].Rows); rowIdx++ {
		curInst := e.rc[0].Rows[rowIdx].File
		gName, ok := e.groupTo[curInst]
		if !ok {
			err = fmt.Errorf("The instance-file %s has no group info", curInst)
			return
		}
		gIdx, ok := e.groupIdx[gName]
		if !ok {
			panic("group in groupTo but not found in groupIdx")
		}
		err = e.procOneRow(rowIdx, pCounting[gIdx])
		if err != nil {
			return
		}
	}

	// summary final
	fCounting = make([]basicCountingCell, len(e.rc))
	init(fCounting)
	for gIdx := range e.groupName {
		basicCountingCellUpdates(pCounting[gIdx][1:])
		basicCountingCellMarks(pCounting[gIdx][1:])
		for rIdx := range e.rc {
			fCounting[rIdx].addFrom(pCounting[gIdx][rIdx])
		}
	}
	basicCountingCellUpdates(fCounting[1:])
	basicCountingCellMarks(fCounting[1:])

	return
}

func (e *sgEngine) Run(writer io.Writer, rc []*report.Report) error {
	if err := e.init(rc); err != nil {
		return err
	}

	type normalPhaseStruct struct {
		InstanceFileCount int
		ReportHeader      []basicReportHeaderCell
		GroupName         []string // name of the groups.
		GroupTo           map[string]string
		GroupCap          []int                 // how many instances are in the group
		PCounting         [][]basicCountingCell // partial counting. len(GroupCounting)==len(GroupName)==len(GroupCap). [group index][report index]
		FCounting         []basicCountingCell   // the whole counting
	}

	table := normalPhaseStruct{
		InstanceFileCount: len(rc[0].Rows),
		ReportHeader:      e.makeReportHeader(),
		GroupName:         e.groupName,
		GroupTo:           e.groupTo,
		GroupCap:          e.groupCap,
	}

	var err error
	table.PCounting, table.FCounting, err = e.makeCounting()
	if err != nil {
		return err
	}

	return e.theTemplate.Execute(writer, table)
}

func (e *sgEngine) Name() string { return "sg" }

func (e *sgEngine) AdditionalHelp() string {
	return "Compare separately and group by given rules."
}

func (e *sgEngine) SetFlags(flag *flag.FlagSet) {
	e.basicEngine.SetFlags(flag)
	flag.StringVar(&e.groupFile, "group", "", "The grouping rules should be provided in `FILE`.")
}

func init() {
	addEngine(new(sgEngine))
}
