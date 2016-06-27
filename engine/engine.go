package engine

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/femrat/rcmp/report"
	"github.com/femrat/rcmp/stderr"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var myBaseDir string

func SetMyBaseDir(dir string) {
	myBaseDir = dir
}

type Engine interface {
	// parse the arguments. DON'T DO OTHER THINGS HERE.
	SetFlags(*flag.FlagSet)

	// start engine
	Run(io.Writer, []*report.Report) error

	// additional help after flags
	AdditionalHelp() string

	// get engine name
	Name() string
}

var engineMap = make(map[string]Engine)

func GetEngine(engineName string) Engine {
	return engineMap[engineName]
}

func PrintEngineHelp(out io.Writer) {
	for name, engine := range engineMap {
		stderr.Logf("[engine: %s]: %s\n", name, engine.AdditionalHelp())
		set := new(flag.FlagSet)
		engine.SetFlags(set)
		set.SetOutput(out)
		set.PrintDefaults()
		stderr.Logln()
	}
}

func addEngine(engine Engine) {
	if _, ok := engineMap[engine.Name()]; ok {
		panic(fmt.Sprintf("engine exists: %s", engine.Name()))
	}
	engineMap[engine.Name()] = engine
}

func alignString(splitStr, newSplitStr, inStr string) string {
	if splitStr == "" {
		panic("need splitStr")
	}

	// calculate the length of text on screen, control chars are not counted.
	// Three valid control form:
	// \033[1m
	// \033[1;31m
	// \033[1;31;43m
	// If the control sequence is not valid, like "\033[1;31;m", all sequence is still invisible.
	r := regexp.MustCompile("\033\\[[0-9;]*m")
	plainLen := func(str string) int {
		return len(r.ReplaceAllString(str, ""))
	}
	_ = plainLen("A")

	var table [][]string
	var terminateWithSlashN bool

	maxColNumInRow := 0

	bi := bufio.NewReader(bytes.NewBufferString(inStr))
	for {
		line, err := bi.ReadString('\n')

		if err == io.EOF {
			if line == "" {
				terminateWithSlashN = true
				break
			} else {
				terminateWithSlashN = false
			}
		}

		line = strings.TrimRight(line, "\r\n")
		s := strings.Split(line, splitStr)
		table = append(table, s)
		if maxColNumInRow < len(s) {
			maxColNumInRow = len(s)
		}
		if err == io.EOF {
			break
		}
	}

	for colIdx := 0; colIdx < maxColNumInRow; colIdx++ {
		maxLen := 0
		for _, row := range table {
			// note: len(row)-1, pass the last element in the row, allow the last cell to overflow
			if len(row)-1 > colIdx {
				curLen := plainLen(row[colIdx])
				if maxLen < curLen {
					maxLen = curLen
				}
			}
		}
		for _, row := range table {
			if len(row)-1 > colIdx {
				row[colIdx] += strings.Repeat(" ", maxLen-plainLen(row[colIdx]))
			}
		}
	}

	out := new(bytes.Buffer)
	for i, row := range table {
		fmt.Fprintf(out, strings.Join(row, newSplitStr))
		if i < len(table)-1 || terminateWithSlashN {
			fmt.Fprintln(out)
		}
	}
	return out.String()
}

func addTemplateFuncs(tpl *template.Template) {
	tpl.Funcs(template.FuncMap{
		"TemplateToString": func(name string, data interface{}) string {
			buf := new(bytes.Buffer)
			if err := tpl.ExecuteTemplate(buf, name, data); err != nil {
				panic(err)
			}
			return buf.String()
		},
		"Color": func(seq ...int) string {
			switch len(seq) {
			case 0:
				return "\033[m"
			case 1:
				return fmt.Sprintf("\033[%dm", seq[0])
			case 2:
				return fmt.Sprintf("\033[%d;3%dm", seq[0], seq[1])
			case 3:
				return fmt.Sprintf("\033[%d;3%d;4%dm", seq[0], seq[1], seq[2])
			default:
				panic("Color function has up to 3 arguments")
			}
		},
		"Align":      func(splitStr, newSplitStr, inStr string) string { return alignString(splitStr, newSplitStr, inStr) },
		"Basename":   func(s string) string { return filepath.Base(s) },
		"TrimSuffix": func(suffix, s string) string { return strings.TrimSuffix(s, suffix) },
		"TrimPrefix": func(prefix, s string) string { return strings.TrimSuffix(s, prefix) },
		"Add":        func(i, j int) int { return j + i },
		"Sub":        func(i, j int) int { return j - i }, // (pipeline | sub i)
		"Replace":    func(old, new, str string) string { return strings.Replace(old, new, str, -1) },
		"Tex": func(str string, env ...string) string {
			if len(env) == 0 {
				str = strings.Replace(str, "\\", "\\textbackslash", -1)
			} else {
				if env[0] != "formula" || len(env) > 1 {
					panic("the only one option env can only be formula now")
				}
				str = strings.Replace(str, "\\", "\\backslash", -1)
			}
			for _, c := range "#$%&_{}[]" {
				str = strings.Replace(str, string(c), "\\"+string(c), -1)
			}
			str = strings.Replace(str, "~", "\\~{}", -1)
			str = strings.Replace(str, "^", "\\^{}", -1)
			return str
		},
	})
}

func getTemplateFromDisk(templateFile string) (string, error) {
	for _, fn := range []string{templateFile, myBaseDir + "/template/" + templateFile} {
		fp, err := os.Open(fn)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return "", err
		}
		defer fp.Close()
		b, err := ioutil.ReadAll(fp)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	return "", fmt.Errorf("template file %s not found", templateFile)
}

func markBest(len int, isBetter func(int, int) int64, mark func(int)) {
	bestIdx := 0
	for i := 1; i < len; i++ {
		if isBetter(i, bestIdx) > 0 {
			bestIdx = i
		}
	}
	for i := 0; i < len; i++ {
		if isBetter(bestIdx, i) == 0 {
			mark(i)
		}
	}
}
