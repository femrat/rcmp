package main

import (
	"github.com/femrat/rcmp/cmp"
	"os"
)

func main() {
	cmp.Go(os.Args[0], os.Args[1:])
}
