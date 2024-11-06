package main

type inf interface {
	call1()
	call2()
}
type struct0 struct {
	a0 int
}

var a123456789 = ""

type struct1 struct {
	*struct0
	a, b, c int
}

func (a *struct1) a3456(c, b int) {
	var s = &struct0{}
}

// a.call2()
// call1 test
func (a struct1) call1() {
	a.a3456(1, 2)
}

// call2  test2
func (a *struct1) call2(c, b int) {
	// a.call2()
	var s = &struct0{}
}
func aa() {
}
func a2() {
}

func c123() {
}
