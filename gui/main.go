package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// root := flag.String("root", "", "root-dir")
	// file := flag.String("file", "", "source file")
	// flag.Parse()
	// var arg = &mainui.Arguments{
	// 	Root: *root,
	// 	File: *file,
	// }
	// mainui.MainUI(arg)

	a := app.New()
	w := a.NewWindow("Hello")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))

	w.ShowAndRun()

}
