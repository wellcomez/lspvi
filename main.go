package main

import (
	"flag"
	"path/filepath"

	// pty "zen108.com/lspvi/pkg/pty"
	mainui "zen108.com/lspvi/pkg/ui"
	"zen108.com/lspvi/pkg/ui/common"
	web "zen108.com/lspvi/pkg/ui/xterm"
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
	var arg = &common.Arguments{
		Root: *root,
		File: *file,
		Tty:  *tty,
		Ws:   *ws,
		Grep: *grep,
		Help: *help,
	}
	if *gui {
		dir := *root
		if dir == "" {
			dir, _ = filepath.Abs(".")
		}
		web.SetPjrRoot(dir)
		web.StartWebUI(*arg, nil)
		return
	}
	mainui.MainUI(arg)
}
