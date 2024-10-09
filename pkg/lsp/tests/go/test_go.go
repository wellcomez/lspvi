package main

type struct1 struct {
	a, b, c int
}

func (a struct1) call1()    {}
func (a *struct1) call2() {}
