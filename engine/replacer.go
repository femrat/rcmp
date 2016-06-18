package engine

import (
	"fmt"
	"io"
	"os"
)

type Replacer map[string]string

func (r *Replacer) Load(in io.Reader) error {
	if *r == nil {
		*r = make(map[string]string)
	}
	var old, new string
	for {
		n, err := fmt.Fscan(in, &old, &new)
		if n == 2 {
			if _, ok := (*r)[old]; ok {
				return fmt.Errorf("Replacer found %s duplicated", old)
			}
			(*r)[old] = new
		} else if n == 0 {
			if err == io.EOF {
				break
			} else {
				return err
			}
		} else {
			return fmt.Errorf("Replacer file should be in pair, [%s] found alone", old)
		}
	}
	return nil
}

func (r *Replacer) LoadFromDisk(filename string) error {
	fp, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	if err = r.Load(fp); err != nil {
		return err
	}
	return nil
}

func (r Replacer) Replace(old string) string {
	new, ok := r[old]
	if ok {
		return new
	} else {
		return old
	}
}
