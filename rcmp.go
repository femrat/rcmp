package main

import (
	"github.com/femrat/rcmp/cmp"
	"os"
	"path/filepath"
)

func main() {
	path := os.Getenv("RCMP_PATH")
	if path == "" {
		path = filepath.Dir(os.Args[0])
	}
	cmp.Go(path, os.Args[1:])
}
