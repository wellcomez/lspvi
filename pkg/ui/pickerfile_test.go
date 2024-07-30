package mainui

import (
	"path/filepath"
	"testing"
)

func TestXxx(t *testing.T) {
	// var task = filewalk{}
	var dir = "/home/z/dev/lsp/goui"
	dir = "/chrome/buildcef/chromium/src"
	ret := &filewalk{
		ret:     []string{},
		root:    dir,
		ignores: WalkerSkip,
	}
	ret.readFiles(ret.root)
	t.Log(len(ret.ret))
}
func TestXxxIgnore(t *testing.T) {
	// var task = filewalk{}
	root:="/chrome/buildcef/chromium/src/third_party"
	a:=NewGitIgnore(root)
	check_dir := filepath.Join(root, "/wlcs/src")
	yes:=a.Ignore(check_dir)
	if yes==false{
		t.Fatal(check_dir)
	}
}
func TestXxxIgnore_doto(t *testing.T) {
	// var task = filewalk{}
	root:="/chrome/buildcef/chromium/src"
	a:=NewGitIgnore(root)
	
	file:="/out/Debug/obj/chrome/browser/browser/navigation_throttle.o"
	check_dir := filepath.Join(root, file)
	check_dir="/chrome/buildcef/chromium/src/out/Debug/obj/chrome/gpu/gpu/chrome_content_gpu_client.o"
	yes:=a.Ignore(check_dir)
	if yes==false{
		t.Fatal(check_dir)
	}
}

func TestXxxIgnore_cscsope(t *testing.T) {
	// var task = filewalk{}
	root:="/chrome/buildcef/chromium/src"
	a:=NewGitIgnore(root)
	check_dir := filepath.Join(root, "/cscope.out")
	yes:=a.Ignore(check_dir)
	if yes==false{
		t.Fatal(check_dir)
	}
}
