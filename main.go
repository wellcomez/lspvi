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
	ws := flag.String("ws", "", "websocket address")
	tty := flag.Bool("tty", false, "guimain")
	grep := flag.Bool("grep", false, "grep")
	help := flag.Bool("help", false, "help")
	root := flag.String("root", "", "root-dir")
	file := flag.String("file", "", "source file")
	flag.Parse()
	var arg = &mainui.Arguments{
		Root: *root,
		File: *file,
		Tty:  *tty,
		Ws:   *ws,
		Grep: *grep,
		Help: *help,
	}
	if *gui {
		mainui.StartWebUI(*arg, nil)
		return
	}
	mainui.MainUI(arg)
}
