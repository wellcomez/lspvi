package mainui

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	// "github.com/gdamore/tcell"
)

type CodeView struct {
	filename string
	view     *femto.View
	main     *mainui
}

func NewCodeView(main *mainui) *CodeView {
	view := tview.NewTextView()
	view.SetBorder(true)
	ret := CodeView{}
	ret.main = main
	var colorscheme femto.Colorscheme
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, "monokai"); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}
	path := ""
	content := ""
	buffer := femto.NewBufferFromString(string(content), path)
	root := femto.NewView(buffer)
	root.SetRuntimeFiles(runtime.Files)
	root.SetColorscheme(colorscheme)

	root.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		// x, y := event.Position()
		// log.Printf("mount action=%d  x=%d y=%d", action, x, y)
		x1, y1, x2, y2 := root.GetInnerRect()
		leftX, _, _, _ := root.GetRect()
		posX, posY := event.Position()
		if posX < x1 || posY > y2 || posY < y1 || posX > x2 {
			return action, event
		}

		log.Print(x1, y1, x2, y2)
		if action == tview.MouseLeftClick {
			// x, y := event.Position()
			posY = posY + root.Topline
			posX = posX - leftX - 3
			root.Cursor.Loc = femto.Loc{X: posX, Y: posY}
			root.Cursor.SetSelectionStart(femto.Loc{X: posX, Y: posY})
			root.Cursor.SetSelectionEnd(femto.Loc{X: posX, Y: posY})
			// root.SelectLine()
			return tview.MouseConsumed, nil
		}
		if action == 14 || action == 13 {
			gap := 2
			if action == 14 {
				posY = posY + gap
				root.ScrollDown(gap)
			} else {
				posY = posY - gap
				root.ScrollUp(gap)
			}
			posX = posX - leftX
			root.Cursor.Loc = femto.Loc{X: posX, Y: femto.Max(0, femto.Min(posY+root.Topline, root.Buf.NumLines))}
			log.Println(root.Cursor.Loc)
			root.SelectLine()
			return tview.MouseConsumed, nil
		}
		return action, event
	})
	root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'r':
			loc := root.Cursor.CurSelection
			ret.main.OnReference(lsp.Range{
				Start: lsp.Position{
					Line:      loc[0].Y,
					Character: loc[0].X,
				},
				End: lsp.Position{
					Line:      loc[1].Y,
					Character: loc[1].X,
				},
			}, main.codeview.filename)
			return nil
		}
		switch event.Key() {
		case tcell.KeyUp:
			root.Buf.LinesNum()
			// root.CursorUp()
			root.SelectUp()
			root.ScrollUp(1)
			root.SelectLine()
			log.Println("cursor up ", root.Cursor.CurSelection[0], root.Cursor.CurSelection[1])
		case tcell.KeyDown:
			root.SelectDown()
			root.ScrollDown(1)
			root.SelectLine()
			// root.SelectLine()
			log.Println("cursor down ", root.Cursor.CurSelection[0], root.Cursor.CurSelection[1])
		case tcell.KeyCtrlS:
			// saveBuffer(buffer, path)
			return nil
		case tcell.KeyCtrlQ:
			return nil
		}
		return nil
	})
	ret.view = root
	return &ret
}
func (code *CodeView) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	buffer := femto.NewBufferFromString(string(data), filename)
	code.view.OpenBuffer(buffer)
	code.filename = filename

	var colorscheme femto.Colorscheme
	if monokai := runtime.Files.FindFile(femto.RTColorscheme, "monokai"); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}
	code.view.SetColorscheme(colorscheme)

	code.view.SetTitle(filename)
	return nil
}
func (codeview *CodeView) gotoline(line int) {
	log.Println("gotoline", line)
	codeview.view.Topline = line
	RightX := len(codeview.view.Buf.Line(line))
	codeview.view.Cursor.CurSelection[0] = femto.Loc{
		X: 0,
		Y: line,
	}
	codeview.view.Cursor.CurSelection[0] = femto.Loc{
		X: RightX,
		Y: line,
	}
	root := codeview.view
	root.Cursor.Loc = femto.Loc{X: 0, Y: line}
	root.Cursor.SetSelectionStart(femto.Loc{X: 0, Y: line})
	text := root.Buf.Line(line)
	root.Cursor.SetSelectionEnd(femto.Loc{X: len(text), Y: line})
}
