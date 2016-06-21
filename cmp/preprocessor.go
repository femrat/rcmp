package cmp

import (
	"flag"
	"fmt"
	"github.com/femrat/rcmp/report"
	"github.com/femrat/rcmp/stderr"
	"io"
	"os"
)

type preprocessor struct {
	sDupReport    bool
	sKeepBasename bool
	sTrimSuffix   string
	sFilter       string
	sFilterSort   bool
	sIntersection bool
	sSplitStr     string

	ReportCollection []*report.Report
}

func newPreProcessor() *preprocessor {
	p := new(preprocessor)
	return p
}

func (p *preprocessor) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.sDupReport, "dup-report", false,
		"Accept duplicated reports, instead of ignoring the latter duplicated reports.")
	f.BoolVar(&p.sKeepBasename, "keep-basename", false,
		"Don't strip directory from instance-file in every report line automatically.")
	f.StringVar(&p.sTrimSuffix, "trim-suffix", "",
		"Trim suffix `SUFFIX` from instance-file, such as extension.")
	f.StringVar(&p.sFilter, "filter-file", "",
		"Filter the instance-file by given `FILE`, include only the specificed instance-files.")
	f.BoolVar(&p.sFilterSort, "filter-sort", false,
		"Sort the rows in each report by the order of filter. Must be used with -filter-file.")
	f.BoolVar(&p.sIntersection, "intersection", false,
		"Compare only the intersection of instance-files of the reports.")
	f.StringVar(&p.sSplitStr, "split", "\t",
		"The string to split each element in report line.")
}

func (p *preprocessor) Run(filenames []string) error {
	if p.sFilterSort == true && p.sFilter == "" {
		return fmt.Errorf("The option filter-sort must be used with filter-file.")
	}

	// remove dup report
	if !p.sDupReport {
		filenames = p.removeDupReports(filenames)
	}

	// load files
	for _, fn := range filenames {
		if rep, err := report.NewReportFromDisk(fn, p.sSplitStr); err != nil {
			return fmt.Errorf("Error while loading %s: %v", fn, err)
		} else {
			p.ReportCollection = append(p.ReportCollection, rep)
		}
	}

	// basename
	if !p.sKeepBasename {
		for _, rep := range p.ReportCollection {
			rep.Basename()
		}
	}

	// safety check for dup rows in every report
	if err := p.checkDupRow(); err != nil {
		return err
	}

	// suffix
	if p.sTrimSuffix != "" {
		for _, rep := range p.ReportCollection {
			rep.StripSuffix(p.sTrimSuffix)
		}
	}

	// filter and filter sort
	if p.sFilter != "" {
		fp, err := os.Open(p.sFilter)
		if err != nil {
			return fmt.Errorf("Can't open filter %s: %v", p.sFilter, err)
		}
		defer fp.Close()
		if err := p.filter(fp); err != nil {
			return err
		}
	}

	if p.sFilterSort && p.sFilter == "" {
		return fmt.Errorf("-filter-sort must be used with -filter")
	}

	if p.sIntersection {
		if err := p.intersection(); err != nil {
			return err
		}
	}

	return nil
}

func (p *preprocessor) removeDupReports(filenames []string) []string {
	exist := make(map[string]bool)
	for i := 0; i < len(filenames); {
		if exist[filenames[i]] {
			stderr.Logf("Warning: duplicated report %s removed\n", filenames[i])
			filenames = append(filenames[:i], filenames[i+1:]...)
		} else {
			exist[filenames[i]] = true
			i++
		}
	}
	return filenames
}
func (p *preprocessor) checkDupRow() error {
	for _, rep := range p.ReportCollection {
		m := make(map[string]bool)
		for _, row := range rep.Rows {
			if m[row.File] {
				return fmt.Errorf("Duplicated instance-file %s found in %s", row.File, rep.ReportFile)
			}
			m[row.File] = true
		}
	}
	return nil
}

func (p *preprocessor) filter(fp io.Reader) error {
	fmap := make(map[string]bool)
	var order []string
	for {
		var fn string
		if n, err := fmt.Fscanf(fp, "%s", &fn); n == 1 {
			if fmap[fn] {
				return fmt.Errorf("No duplicated line allowed in filter-file. Duplicate line detected: %s", fn)
			}
			fmap[fn] = true
			order = append(order, fn)
		} else if err == io.EOF {
			break
		} else {
			fmt.Errorf("Error while read filter %s: %v", p.sFilter, err)
		}
	}

	for _, rep := range p.ReportCollection {
		var newRows []*report.ReportRow
		for _, row := range rep.Rows {
			if fmap[row.File] {
				newRows = append(newRows, row)
			}
		}
		if len(rep.Rows)-len(newRows) > 0 {
			stderr.Logf("Warning: report %s shrinked %d rows, left %d rows\n", rep.ReportFile, len(rep.Rows)-len(newRows), len(newRows))
		}
		rep.Rows = newRows

		if p.sFilterSort {
			m := make(map[string]int)
			newRows = nil
			for i, row := range rep.Rows {
				m[row.File] = i
			}
			for _, fn := range order {
				if i, ok := m[fn]; ok {
					newRows = append(newRows, rep.Rows[i])
				}
			}
			if len(newRows) != len(rep.Rows) {
				panic("impossible")
			}
			if len(order)-len(newRows) > 0 {
				stderr.Logf("Warning: report %s sorted but %d rows missed compared to the given filter-file\n", rep.ReportFile, len(order)-len(newRows))
			}
			rep.Rows = newRows
		}
	}

	return nil
}

func (p *preprocessor) intersection() error {
	exist := make(map[string]int)
	for _, rep := range p.ReportCollection {
		for _, row := range rep.Rows {
			exist[row.File]++
		}
	}

	count := 0
	for _, cnt := range exist {
		if cnt == len(p.ReportCollection) {
			count++
		} else if cnt > len(p.ReportCollection) {
			panic("impossible")
		}
	}

	if count == 0 {
		return fmt.Errorf("No instance-file left after intersection operation")
	}

	makeChanged := false
	for _, rep := range p.ReportCollection {
		var newRows []*report.ReportRow
		for _, row := range rep.Rows {
			if exist[row.File] == len(p.ReportCollection) {
				newRows = append(newRows, row)
			}
		}
		if len(newRows) != len(rep.Rows) {
			stderr.Logf("Warning: intersection takes %d rows off from %s\n", len(rep.Rows)-len(newRows), rep.ReportFile)
			rep.Rows = newRows
			makeChanged = true
		}
	}

	if makeChanged {
		stderr.Logf("After intersection operation, %d instance-file left\n", count)
		//} else {
		//stderr.Logf("After intersection operation, no instance-file gets removed, %d instance-file left\n", count)
	}

	return nil
}
