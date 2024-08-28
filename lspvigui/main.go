package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type VSCodeObjectCodec struct{}

// WriteObject implements ObjectCodec.
func (VSCodeObjectCodec) WriteObject(stream io.Writer, obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stream, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
		return err
	}
	if _, err := stream.Write(data); err != nil {
		return err
	}
	return nil
}

// ReadObject implements ObjectCodec.
func (VSCodeObjectCodec) ReadObject(stream *bufio.Reader, v interface{}) error {
	for {
		buf := []byte{}
		n, err := stream.Read(buf)
		if err != nil {
			return err
		}
		if n > 0 {
			log.Println("readed", n, err)
		}
	}
}

// object in a stream.
type ObjectCodec interface {
	// WriteObject writes a JSON-RPC 2.0 object to the stream.
	WriteObject(stream io.Writer, obj interface{}) error

	// ReadObject reads the next JSON-RPC 2.0 object from the stream
	// and stores it in the value pointed to by v.
	ReadObject(stream *bufio.Reader, v interface{}) error
}

type bufferedObjectStream struct {
	conn io.Closer // all writes should go through w, all reads through r
	w    *bufio.Writer
	r    *bufio.Reader

	codec ObjectCodec

	mu sync.Mutex
}

func (t *bufferedObjectStream) WriteObject(obj interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := t.codec.WriteObject(t.w, obj); err != nil {
		return err
	}
	return t.w.Flush()
}

// ReadObject implements ObjectStream.
func (t *bufferedObjectStream) ReadObject(v interface{}) error {
	return t.codec.ReadObject(t.r, v)
}

// Close implements ObjectStream.
func (t *bufferedObjectStream) Close() error {
	return t.conn.Close()
}


const ioctlSetControllingTty = 0x5412 // TIOCSCTTY
func run_term() (*bufferedObjectStream, error) {
	maintty()
	root := filepath.Dir(os.Args[0])
	relpath := filepath.Join(root, "..", "lspvi")
	lspvi, err := filepath.Abs(relpath)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(lspvi, "--root", filepath.Dir(lspvi))
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err

	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	conn := struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: stdout,
		Writer: stdin,
		Closer: stdin,
	}
	ss := &bufferedObjectStream{
		conn:  conn,
		w:     bufio.NewWriter(conn),
		r:     bufio.NewReader(conn),
		codec: VSCodeObjectCodec{},
	}
	go ss.ReadObject(nil)
	return ss, nil

}

func main() {
	// root := flag.String("root", "", "root-dir")
	// file := flag.String("file", "", "source file")
	// flag.Parse()
	// var arg = &mainui.Arguments{
	// 	Root: *root,
	// 	File: *file,
	// }
	// mainui.MainUI(arg)

	a := app.New()
	w := a.NewWindow("Hello")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))
	run_term()
	w.ShowAndRun()

}
