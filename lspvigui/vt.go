package main

import (
	"io"
	// "log"
	"os"
	"os/exec"

	"github.com/creack/pty"
)
type read_out struct{

}
func main() {
	// 创建一个新的伪终端
	lspvi := "/Users/jialaizhu/dev/lspvi/lspvi"
	c:=exec.Command(lspvi)
	f, err := pty.Start(c)
	if err != nil {
		panic(err)
	}
	var stdout2 read_out
	io.Copy(os.Stdout, f)
	// EOT
	// newFunction()
}

func newFunction() {
	c := exec.Command("grep", "--color=auto", "bar")
	f, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	go func() {
		f.Write([]byte("foo\n"))
		f.Write([]byte("bar\n"))
		f.Write([]byte("baz\n"))
		f.Write([]byte{4})
	}()
	io.Copy(os.Stdout, f)
}
