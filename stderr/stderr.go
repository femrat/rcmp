package stderr

import (
	"fmt"
	"os"
)

func Logf(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, format, values...)
}

func Logln(values ...interface{}) {
	fmt.Fprintln(os.Stderr, values...)
}
