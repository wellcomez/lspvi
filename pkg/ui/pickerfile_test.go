package mainui

import "testing"

func TestXxx(t *testing.T) {
	var task = filewalk{}
	var dir = "/home/z/dev/lsp/goui"
	dir = "/chrome/buildcef/chromium/src"
	task.readFiles(dir, WalkerSkip)
	t.Log("lenght", len(task.ret))
}
