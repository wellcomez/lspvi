package grep

import "testing"

func Test_grep(t *testing.T) {
	pattern := "mainui"
	opt := OptionSet{
		Grep_only: true,
		G:         true,
	}
	g, err := NewGorep(1, pattern, &opt)
	if err != nil {
		t.Fatal(err)
	}
	g.CB = func(taskid int, out *GrepOutput) {
		t.Log(out.Line, out.LineNumber, out.Fpath)
	}
	fpath := "/Users/jialaizhu/dev/lspgo"
	chans := g.kick(fpath)
	g.report(chans, false)
}
