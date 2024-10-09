package main

type struct1 struct {
	a, b, c int
}

func (struct1) call1()  {}
func (*struct1) call2() {}
