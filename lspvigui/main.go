package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	// "golang.org/x/tools/go/analysis/passes/printf"
	pty "zen108.com/lspvi/pkg/pty"
)

type guio struct {
}

// write implements pty.ptyio.
func (g guio) Write(s string) {
	// panic("unimplemented")
	// println(s)
	// var text []fyne.CanvasObject
	for _, v := range s {
		// printf(" %c",v)	
		if v == '\n' {
			fmt.Println()
			continue
		}
		fmt.Printf(" %c",v)
	}
}


func main() {
	myApp := app.New()
	w := myApp.NewWindow("Text")

	text := canvas.NewText("Text Object", color.Black)
	text.Alignment = fyne.TextAlignTrailing
	text.TextStyle = fyne.TextStyle{Italic: true}
	var g guio
	go func ()  {
		pty.Ptymain([]string{"/usr/bin/lspvi"}, g)
	}()
	w.SetContent(text)
	w.ShowAndRun()
}
