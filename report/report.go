package report

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ReportRow stands for one line in a report file
type ReportRow struct {
	File   string
	Values []string
}

// Report stands for a report file
type Report struct {
	ReportFile string
	Rows       []*ReportRow
}

func (r *Report) Basename() {
	for i := range r.Rows {
		r.Rows[i].File = filepath.Base(r.Rows[i].File)
	}
}

func (r *Report) StripSuffix(ext string) {
	// Don't strip extension here, since some instances name may have dot originally
	// A specified extension is needed.
	for i := range r.Rows {
		r.Rows[i].File = strings.TrimRight(r.Rows[i].File, ext)
	}
}

func NewReport(reader io.Reader, fileName, splitStr string) (*Report, error) {
	rep := new(Report)
	rep.ReportFile = fileName
	br := bufio.NewReader(reader)
	for {
		line, err := br.ReadString('\n')
		if err == nil {
			line = strings.TrimRight(line, "\r\n")
			if len(line) == 0 {
				return nil, fmt.Errorf("Meet an empty line in %s", fileName)
			}
			row := new(ReportRow)
			slice := strings.Split(line, splitStr)
			row.File = slice[0]
			row.Values = slice[1:]
			rep.Rows = append(rep.Rows, row)
		} else if err == io.EOF {
			break
		} else {
			return nil, err
		}
	}
	return rep, nil
}

func NewReportFromDisk(fileName, splitStr string) (*Report, error) {
	fp, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	r, err := NewReport(fp, fileName, splitStr)
	if err != nil {
		return nil, err
	}

	return r, nil
}
