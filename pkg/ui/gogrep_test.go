package mainui

import "testing"
func Test_grep(t *testing.T) {
	pattern :="fzf"
	opt:=optionSet{
		// grep_only:true,
		g:true,

	}
	g ,err:= newGorep(1,pattern, &opt)
	if err != nil {
		t.Fatal(err)
	}
	fpath :="/Users/jialaizhu/dev/lspgo"
	chans := g.kick(fpath)
	g.report(chans, verifyColor())
}