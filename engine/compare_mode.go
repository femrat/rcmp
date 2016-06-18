package engine

import "fmt"

type compareFunc struct {
	// non-valid opt SHOULD NOT BE COMPARED
	IsValid func(a int64) bool

	// use ONLY IF both a and b are valid
	// >0: a is better than b. ==0: equal. <0 a is worse than b
	IsBetter func(a, b int64) int64
}

func newCompareFunc(mode string) (*compareFunc, error) {
	c := new(compareFunc)
	switch mode {
	case "sat":
		c.IsValid = func(a int64) bool { return a >= 0 }
		c.IsBetter = func(a, b int64) int64 { return b - a } // >0 means a is better
	}

	if c.IsBetter != nil || c.IsValid != nil {
		return c, nil
	} else {
		return nil, fmt.Errorf("compareMode not support")
	}
}
