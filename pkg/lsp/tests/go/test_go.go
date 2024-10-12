package main

type inf interface {
	call1()
	call2()
}
type struct0 struct {
	a0 int
}
type struct1 struct {
	*struct0
	a, b, c int
}

func (a struct1) call1()  {}
func (a *struct1) call2() {}
func aa(){
}
func aaaaa(){
}
