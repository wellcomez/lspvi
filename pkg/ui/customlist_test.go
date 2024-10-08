package mainui

import (
	"log"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func Test_mainui_Init(t *testing.T) {
	sss := fmt_color_string("123", tcell.ColorDarkGray)
	x1 := "123" + sss + "456"
	s, c,pos := pasrse_color_string(x1)
	if s != "123" {
		t.Error("s!=123")
	}
	if c != tcell.ColorDarkGray {
		t.Error("c!=tcell.ColorDarkGray")
	}
	x := x1[pos.X:pos.Y]
	log.Println(x)
}
