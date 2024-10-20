package main

import (
	"flag"

	// pty "zen108.com/lspvi/pkg/pty"
	mainui "zen108.com/lspvi/pkg/ui"
)

type arg struct {
	count int
	ws    string
}

func main() {
	gui := flag.Bool("gui", false, "guimain")
	ws := flag.String("ws", "", "guimain")
	tty := flag.Bool("tty", false, "guimain")
	root := flag.String("root", "", "root-dir")
	file := flag.String("file", "", "source file")
	flag.Parse()
	var arg = &mainui.Arguments{
		Root: *root,
		File: *file,
		Tty:  *tty,
		Ws:   *ws,
	}
	if *gui {
		mainui.StartWebUI(*arg,nil)
		return
	}
	mainui.MainUI(arg)
}
