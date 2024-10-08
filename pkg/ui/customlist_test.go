package mainui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func Test_mainui_Init(t *testing.T) {
	sss := fmt_color_string("abc", tcell.ColorDarkGray)
	x1 := "123" + sss + "456"
	r0 := pasrse_color_string(x1)
	if r0.m.text != "abc" {
		t.Error("s!=123")
	}
	if r0.m.color != tcell.ColorDarkGray {
		t.Error("c!=tcell.ColorDarkGray")
	}
	if r0.a.text != "456" {
		t.Error("e!=456")
	}
	if r0.b.text != "123" {
		t.Error("b!=123")
	}
	r := pasrse_bold_color_string("a**123**b")
	println(r.b.text, r.m.text, r.a.text)
}
