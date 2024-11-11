package main

import (
	"zen108.com/lspvi/pkg/ptyproxy"
)
func main(){
	ptyproxy.NewMain([]string{"bash"})
}