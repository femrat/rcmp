package cmp

import (
	"bytes"
	"fmt"
	"github.com/femrat/rcmp/report"
	"github.com/femrat/rcmp/stderr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPreProcessor(t *testing.T) {
	assert.NotNil(t, newPreProcessor())
}

func TestRemoveDupReports(t *testing.T) {
	p := newPreProcessor()
	//stderr.SetOutput(nil)
	chk := func(in, expect, warn []string) {
		buf := new(bytes.Buffer)
		stderr.SetOutput(buf)
		assert.Equal(t, expect, p.removeDupReports(in))
		var wstr = ""
		for _, w := range warn {
			wstr += fmt.Sprintf("Warning: duplicated report %s removed\n", w)
		}
		assert.Equal(t, wstr, string(buf.Bytes()))
	}
	chk(nil, nil, nil)
	chk([]string{}, []string{}, nil)
	chk([]string{"a", "b", "c"}, []string{"a", "b", "c"}, nil)
	chk([]string{"a", "b", "b"}, []string{"a", "b"}, []string{"b"})
	chk([]string{"a", "a", "a", "b", "b"}, []string{"a", "b"}, []string{"a", "a", "b"})
	chk([]string{"a", "a", "a", "b", "b", "a"}, []string{"a", "b"}, []string{"a", "a", "b", "a"})
}

func row(instName string, v ...string) *report.ReportRow {
	return &report.ReportRow{File: instName, Values: v}
}

func makeReport(file string, r ...*report.ReportRow) *report.Report {
	return &report.Report{ReportFile: file, Rows: r}
}

func TestCheckDupRow(t *testing.T) {
	p := newPreProcessor()

	p.ReportCollection = []*report.Report{
		makeReport("REPFILE1", row("A", "11"), row("B", "12")),
		makeReport("REPFILE2", row("A", "21"), row("B", "22")),
	}
	assert.NoError(t, p.checkDupRow())

	p.ReportCollection = []*report.Report{
		makeReport("REPFILE1", row("A", "11"), row("B", "12")),
		makeReport("REPFILE2", row("A", "21"), row("A", "22")),
	}
	assert.EqualError(t, p.checkDupRow(), "Duplicated instance-file A found in REPFILE2")

	p.ReportCollection = []*report.Report{
		makeReport("REPFILE1", row("B", "11"), row("B", "12")),
		makeReport("REPFILE2", row("A", "21"), row("A", "22")),
	}
	assert.EqualError(t, p.checkDupRow(), "Duplicated instance-file B found in REPFILE1")
}

func TestFilter(t *testing.T) {
	p := newPreProcessor()

	// test read file
	buf := new(bytes.Buffer)
	buf.WriteString("A\nB\nB")
	assert.EqualError(t, p.filter(buf), "No duplicated line allowed in filter-file. Duplicate line detected: B")

	buf.Reset()
	buf.WriteString("A\nB\nB\n")
	assert.EqualError(t, p.filter(buf), "No duplicated line allowed in filter-file. Duplicate line detected: B")

	buf.Reset()
	buf.WriteString("A\nB B\n")
	assert.EqualError(t, p.filter(buf), "No duplicated line allowed in filter-file. Duplicate line detected: B")

	// test filter
	out := new(bytes.Buffer)
	stderr.SetOutput(out)

	buf.Reset()
	out.Reset()
	buf.WriteString("A\nC\nE\n")
	p.ReportCollection = []*report.Report{
		makeReport("REPFILE1", row("C", "13"), row("A", "11"), row("B", "12"), row("D")),
		makeReport("REPFILE2", row("A", "21"), row("E"), row("B", "22"), row("C", "23"), row("F")),
	}
	if assert.NoError(t, p.filter(buf)) {
		expect := []*report.Report{
			makeReport("REPFILE1", row("C", "13"), row("A", "11")),
			makeReport("REPFILE2", row("A", "21"), row("E"), row("C", "23")),
		}
		assert.Equal(t, expect, p.ReportCollection)
		assert.Equal(t, `Warning: report REPFILE1 shrinked 2 rows, left 2 rows
Warning: report REPFILE2 shrinked 2 rows, left 3 rows
`, string(out.Bytes()))
	}

	// test filter with sort
	p.sFilterSort = true
	buf.Reset()
	out.Reset()
	buf.WriteString("A C E G")
	p.ReportCollection = []*report.Report{
		makeReport("REPFILE1", row("C", "13"), row("A", "11"), row("B", "12"), row("D")),
		makeReport("REPFILE3", row("A", "31"), row("E"), row("G", "37"), row("C", "33")),
		makeReport("REPFILE2", row("A", "21"), row("E"), row("B", "22"), row("C", "23"), row("F")),
		makeReport("REPFILE4", row("A", "41"), row("C", "43", "43a"), row("E"), row("G", "47")),
	}
	if assert.NoError(t, p.filter(buf)) {
		expect := []*report.Report{
			makeReport("REPFILE1", row("A", "11"), row("C", "13")),
			makeReport("REPFILE3", row("A", "31"), row("C", "33"), row("E"), row("G", "37")),
			makeReport("REPFILE2", row("A", "21"), row("C", "23"), row("E")),
			makeReport("REPFILE4", row("A", "41"), row("C", "43", "43a"), row("E"), row("G", "47")),
		}
		assert.Equal(t, expect, p.ReportCollection)

		assert.Equal(t, `Warning: report REPFILE1 shrinked 2 rows, left 2 rows
Warning: report REPFILE1 sorted but 2 rows missed compared to the given filter-file
Warning: report REPFILE2 shrinked 2 rows, left 3 rows
Warning: report REPFILE2 sorted but 1 rows missed compared to the given filter-file
`, string(out.Bytes()))
	}
}

func TestIntersection(t *testing.T) {
	p := newPreProcessor()
	p.ReportCollection = []*report.Report{
		makeReport("REPFILE1", row("C", "13"), row("A", "11"), row("B", "12"), row("D")),
		makeReport("REPFILE3", row("A", "31"), row("E"), row("G", "37"), row("C", "33")),
		makeReport("REPFILE2", row("A", "21"), row("E"), row("B", "22"), row("C", "23"), row("F")),
		makeReport("REPFILE4", row("A", "41"), row("C", "43", "43a"), row("E"), row("G", "47")),
	}

	expect := []*report.Report{
		makeReport("REPFILE1", row("C", "13"), row("A", "11")),
		makeReport("REPFILE3", row("A", "31"), row("C", "33")),
		makeReport("REPFILE2", row("A", "21"), row("C", "23")),
		makeReport("REPFILE4", row("A", "41"), row("C", "43", "43a")),
	}

	out := new(bytes.Buffer)
	stderr.SetOutput(out)

	if assert.NoError(t, p.intersection()) {
		p.intersection()
		assert.Equal(t, expect, p.ReportCollection)
		assert.Equal(t, `Warning: intersection takes 2 rows off from REPFILE1
Warning: intersection takes 2 rows off from REPFILE3
Warning: intersection takes 3 rows off from REPFILE2
Warning: intersection takes 2 rows off from REPFILE4
After intersection operation, 2 instance-file left
`, string(out.Bytes()))
	}

	p.ReportCollection = []*report.Report{
		makeReport("REPFILE1", row("C", "13"), row("A", "11"), row("B", "12"), row("D")),
		makeReport("REPFILE3", row("A", "31"), row("E"), row("G", "37"), row("C", "33")),
		makeReport("REPFILE2", row("E"), row("B", "22"), row("F")),
		makeReport("REPFILE4", row("A", "41"), row("C", "43", "43a"), row("E"), row("G", "47")),
	}
	assert.EqualError(t, p.intersection(), "No instance-file left after intersection operation")
}
