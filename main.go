package main

import (
	"flag"

	// pty "zen108.com/lspvi/pkg/pty"
	mainui "zen108.com/lspvi/pkg/ui"
)

func main() {
	gui := flag.Bool("gui", false, "guimain")
	tty := flag.Bool("tty", false, "guimain")
	root := flag.String("root", "", "root-dir")
	file := flag.String("file", "", "source file")
	flag.Parse()
	if *gui {
		mainui.StartWebUI()
		return
	}
	var arg = &mainui.Arguments{
		Root: *root,
		File: *file,
		Tty:  *tty,
	}
	mainui.MainUI(arg)
}
