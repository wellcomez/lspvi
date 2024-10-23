package mainui

import (
	"github.com/pgavlin/femto"
	"github.com/tectiv3/go-lsp"
	"zen108.com/lspvi/pkg/debug"
)
func format0(ret []lsp.TextEdit, code *CodeView) {
	for y, v := range ret {
		if v.NewText == "\n" {
			continue
		}
		start := femto.Loc{
			X: v.Range.Start.Character,
			Y: v.Range.Start.Line,
		}
		end := femto.Loc{
			X: v.Range.End.Character,
			Y: v.Range.End.Line,
		}
		startline := code.view.Buf.Line(start.Y)
		endline := code.view.Buf.Line(end.Y)
		debug.DebugLog("format", y, ">", []rune(v.NewText), "<", start.Y, ":", start.X, "->", end.Y, ":", end.X)
		debug.DebugLog("format", y, "before-start", startline, len(startline))
		if end.Y > start.Y {
			debug.DebugLog("format", y, "before-end", endline, len(endline))
		}
		code.view.Buf.Replace(start, end, v.NewText)
		debug.DebugLog("format", y, "after       ", code.view.Buf.Line(start.Y))
	}
}
func format2(ret []lsp.TextEdit, code *CodeView) {
	for i := range ret {
		y := len(ret) - i - 1
		v := ret[y]
		if v.NewText == "\n" {

		}
		start := femto.Loc{
			X: v.Range.Start.Character,
			Y: v.Range.Start.Line,
		}
		end := femto.Loc{
			X: v.Range.End.Character,
			Y: v.Range.End.Line,
		}
		startline := code.view.Buf.Line(start.Y)
		endline := code.view.Buf.Line(end.Y)
		debug.DebugLog("format", y, ">", []rune(v.NewText), "<", start.Y, ":", start.X, "->", end.Y, ":", end.X)
		debug.DebugLog("format", y, "before-start", startline, len(startline))
		yes := end.Y > start.Y
		if yes {
			continue
		}
		code.view.Buf.Replace(start, end, v.NewText)
		x := code.view.Buf.Line(start.Y)
		debug.DebugLog("format", y, "after       ", x, len(x))
		if yes {
			debug.DebugLog("format", y, "end       ", endline, len(endline))
			x := code.view.Buf.Line(end.Y)
			debug.DebugLog("format", y, "end-after ", x, len(x))
		}
	}
}
