package lspcore_test

import (
	"testing"

	lspcore "zen108.com/lspvi/pkg/lsp"
)

func Test_complete(t *testing.T) {
	a := "call2(${1:})"
	code := lspcore.NewCompleteCode(a)
	ss, _ := code.Token(0)
	if ss.Text != "call2(" {
		t.Error("call2")
	}
	t.Log("ok")
	a = "call2(${1:},${2:})"
	ss, _ = code.Token(0)
	if ss.Text != "call2(" {
		t.Error("call2")
	}
	code = lspcore.NewCompleteCode(a)
	if code.Len() != 5 {
		t.Error("token!=5")
	}
	code.Text()
	t.Log("ok")
}
