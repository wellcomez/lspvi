package mainui

import (
	"log"
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
	r3 := parse_key_string("ab123cd", "123")
	log.Println(r3)
	s := colorpaser{data: "a123**123**a**[123]abc**"}
	ret := s.Parse()
	x := color_maintext(ret)
	log.Println(x)
	log.Println(ret)
	if ret[0].color==0{
		ss:=colorpaser{data:ret[0].text}
		data:= ss.ParseKey([]string{"123"})
		log.Println(data)
	}
}
