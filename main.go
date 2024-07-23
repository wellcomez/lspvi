package main

import (
	"flag"

	mainui "zen108.com/lspui/pkg/ui"
)

func main() {
	root:=flag.String("root","","root-dir")
	file:=flag.String("file","","source file")
	flag.Parse()
	var arg = &mainui.Arguments{
		Root: *root,
		File: *file,
	}
	mainui.MainUI(arg)
}
