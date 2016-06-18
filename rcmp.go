package main

import (
	"github.com/femrat/rcmp/cmp"
	"os"
)

func main() {

	//type S struct {
	//	S1, S2 string
	//}
	//
	//t := template.New("top")
	//
	//conv := func(tName string, data interface{}) string {
	//	buf := new(bytes.Buffer)
	//	t.ExecuteTemplate(buf, tName, nil)
	//	return ">>" + buf.String() + "<< ++" + fmt.Sprintf("%#v", data) + "++"
	//}
	//t.Funcs(template.FuncMap{"conv": conv})
	//
	//tstr := `{{define "T"}} [this is T] {{end -}} 12345 {{conv "T" $}}`
	//
	//template.Must(t.Parse(tstr))
	//
	//err := t.Execute(os.Stdout, S{"111", "222"})
	//fmt.Printf("\n\nerr: %v\n", err)
	//
	//return

	cmp.Go(os.Args[0], os.Args[1:])
}
