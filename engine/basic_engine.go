package engine

import (
	"flag"
	"fmt"
	"github.com/femrat/rcmp/report"
	"text/template"
)

type basicEngine struct {
	templateFile string
	theTemplate  *template.Template
	compareMode  string
	funcs        *compareFunc

	rc []*report.Report
}

func (b *basicEngine) loadTemplate(defaultTemplate string) error {
	if b.templateFile != "" {
		if theTemplate, err := loadTemplateFromDisk(b.templateFile); err != nil {
			return err
		} else {
			b.theTemplate = theTemplate
		}
	} else {
		if defaultTemplate == "" {
			return fmt.Errorf("This engine does not have a default template. You must assign a template.")
		}
		b.theTemplate = template.New("top")
		template.Must(b.theTemplate.Parse(defaultTemplate))
	}
	return nil
}

func (b *basicEngine) SetFlags(flag *flag.FlagSet) {
	flag.StringVar(&b.templateFile, "t", "", "The template `FILE`. Omit to use the internal template.")
	flag.StringVar(&b.compareMode, "mode", "sat", "The compare mode. You can choose \"sat\" only for now.")
}

func (b *basicEngine) parseCompareFunc() error {
	if fun, err := newCompareFunc(b.compareMode); err != nil {
		return err
	} else {
		b.funcs = fun
		return nil
	}
}

func (b *basicEngine) makeReportHeader() []basicReportHeaderCell {
	hdr := make([]basicReportHeaderCell, len(b.rc))
	for i, rep := range b.rc {
		hdr[i].FileName = rep.ReportFile
	}
	return hdr
}

func (b *basicEngine) makeInstanceFile(replacer func(old string) string) []basicInstanceFileCell {
	inst := make([]basicInstanceFileCell, len(b.rc[0].Rows))
	for i, row := range b.rc[0].Rows {
		if replacer != nil {
			inst[i].FileName = replacer(row.File)
		} else {
			inst[i].FileName = row.File
		}
	}
	return inst
}

// structures ...

type basicCountingValue struct {
	IsGreatest bool // just Greatest, not best (for example, less worse is better)
	IsLeast    bool // numeric compare
	I          int
}

type basicInstanceFileCell struct {
	FileName string
}

type basicReportHeaderCell struct {
	FileName string
}

type basicCountingCell map[string]*basicCountingValue

func (basicCountingCell) Keys() []string {
	return []string{"Better", "Worse", "Delta", "Equal", "EqualWithNA", "BothNA", "NA"}
}

func (s *basicCountingCell) init() {
	if *s != nil {
		panic("double init")
	}
	*s = make(map[string]*basicCountingValue)
	for _, key := range s.Keys() {
		(*s)[key] = new(basicCountingValue)
	}
}

func (s basicCountingCell) addFrom(another basicCountingCell) {
	if len(s) != len(another) {
		panic("map length is not match")
	}
	for _, key := range s.Keys() {
		s[key].I += another[key].I
	}
}

func (s basicCountingCell) exist(key string) basicCountingCell {
	if _, ok := s[key]; !ok {
		panic(fmt.Sprintf("key [%s] not found", key))
	}
	return s
}

func (s basicCountingCell) increase(key string) {
	s.exist(key)
	s[key].I++
}

func basicCountingCellUpdates(counting []basicCountingCell) {
	for i := range counting {
		counting[i]["Delta"].I = counting[i]["Better"].I - counting[i]["Worse"].I
	}
}

func basicCountingCellMarks(counting []basicCountingCell) {
	for _, key := range counting[0].Keys() {
		markBest(len(counting), func(i, j int) int64 { return int64(counting[i][key].I - counting[j][key].I) }, func(k int) { counting[k][key].IsGreatest = true })
		markBest(len(counting), func(i, j int) int64 { return int64(counting[j][key].I - counting[i][key].I) }, func(k int) { counting[k][key].IsLeast = true })
	}
}
