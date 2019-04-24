package main

import (
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {
	message.SetString(language.Chinese, "%s went to %s", "%s去了%s")
}

func main() {
	p := message.NewPrinter(language.Chinese)
	p.Printf("%s went to %s", "撒", "广州")
	fmt.Println()
	//a := []string{"a"}
	//a = append(a, "b")
	//fmt.Println(a)
	//s := "zzasadcads"
	//sList := strings.SplitN(s, "a", 3)
	//fmt.Println(sList)
	//fmt.Println(fmt.Sprintf("zxc %s asd", ""))

	m := &my{}
	Get(m)

}

func Get(m myinter) {
	fmt.Println("name:", m.Getname())
}

type myinter interface {
	Getname() string
}

type my struct {
}

func (self *my) Getname() string {
	return "your name"
}
