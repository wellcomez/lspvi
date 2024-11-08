package ptyproxy

import (
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)
func test_grep() {
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