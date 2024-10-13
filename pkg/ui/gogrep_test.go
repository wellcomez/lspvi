package mainui

import "testing"

func Test_grep(t *testing.T) {
	pattern := "mainui"
	opt := optionSet{
		grep_only: true,
		g:         true,
	}
	g, err := newGorep(1, pattern, &opt)
	if err != nil {
		t.Fatal(err)
	}
	g.cb = func(taskid int, out *grep_output) {
		t.Log(out.Line, out.LineNumber, out.Fpath)
	}
	fpath := "/Users/jialaizhu/dev/lspgo"
	chans := g.kick(fpath)
	g.report(chans, verifyColor())
}
