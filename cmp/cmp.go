package cmp

import (
	"flag"
	"github.com/femrat/rcmp/engine"
	"github.com/femrat/rcmp/stderr"
	"os"
	"path/filepath"
)

var preproc = newPreProcessor()

func printHelpWithPreprocessorDefaults() {
	stderr.Logln(`The rcmp is a simple tool designed for comparison, written by femrat, licenced in GPLv3.
The project can be found in https://github.com/femrat/rcmp.

Basic dataflow:
input(reports) --> [preprocessor] --> [compare engine] --> [template] --> output(table)

The usage of this program is: rcmp ENGINE [options] report1 report2 ...
The input is a series of reports. Each line of the report should follow the format: instance-file value1 value2...
The options may vary due to the choice of engine.

The preprocessor executing chain is 1) dup-check, 2) basename, 3) suffix, 4) filter, 5) filter-sort, 6) intersection. Options are shown as follows.`)

	flag.PrintDefaults()

	stderr.Logln(`
There're engines you can choose. The options of them are listed as follows.
`)

	engine.PrintEngineHelp(os.Stderr)
}

func Go(programName string, args []string) {
	engine.SetMyBaseDir(filepath.Dir(programName))
	preproc.SetFlags(flag.CommandLine)

	if len(args) == 0 {
		// engine name is not optional
		// print "need engine name", and options of preprocessor
		printHelpWithPreprocessorDefaults()
		return
	}

	engine := engine.GetEngine(args[0])
	if engine == nil {
		// print no engine found, and print the list of engines
		stderr.Logln("engine not found")
		return
	}
	args = args[1:]

	engine.SetFlags(flag.CommandLine)

	flag.CommandLine.Parse(args) // preprocessor and engine's options are parsed

	if err := preproc.Run(flag.Args()); err != nil {
		stderr.Logln(err)
		return
	}
	reportCollection := preproc.ReportCollection

	if err := doSafetyCheck(reportCollection); err != nil {
		stderr.Logln(err)
		return
	}

	if err := engine.Run(os.Stdout, reportCollection); err != nil {
		stderr.Logln(err)
		return
	}

}
