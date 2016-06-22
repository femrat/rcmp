package stderr

import (
	"fmt"
	"io"
	"os"
)

var output io.Writer = os.Stderr

func Logf(format string, values ...interface{}) {
	if output != nil {
		fmt.Fprintf(output, format, values...)
	}
}

func Logln(values ...interface{}) {
	if output != nil {
		fmt.Fprintln(output, values...)
	}
}

// set to nil to disable output
func SetOutput(w io.Writer) {
	output = w
}
