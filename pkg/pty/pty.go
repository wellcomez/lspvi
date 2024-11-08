package pty

// https://github.com/iyzyi/aiopty
// https://github.com/photostorm/pty
// https://github.com/creack/pty/pull/155
import (
	"io"
	"log"
	"path/filepath"

	// "log"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"golang.org/x/term"
	"zen108.com/lspvi/pkg/debug"
	// "zen108.com/lspvi/pkg/debug"
)

var home, _ = os.UserHomeDir()
var logFile, _ = setupLogFile(filepath.Join(home, ".lspvi", "ttylogfile.txt"))

var gui io.Writer

type read_out struct {
	handle func(p []byte) (n int, err error)
}

// func test(s string) error {
// 	// Create arbitrary command.
// 	c := exec.Command(s)

// 	// Start the command with a pty.
// 	ptmx, err := pty.Start(c)
// 	if err != nil {
// 		return err
// 	}
// 	// Make sure to close the pty at the end.
// 	defer func() { _ = ptmx.Close() }() // Best effort.

// 	// Handle pty size.
// 	ch := make(chan os.Signal, 1)
// 	signal.Notify(ch, syscall.SIGWINCH)
// 	go func() {
// 		for range ch {
// 			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
// 				debug.ErrorLogf("pty", "error resizing pty: %s", err)
// 			}
// 		}
// 	}()
// 	ch <- syscall.SIGWINCH                        // Initial resize.
// 	defer func() { signal.Stop(ch); close(ch) }() // Cleanup signals when done.

// 	// Set stdin in raw mode.
// 	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

// 	// Copy stdin to the pty and the pty to stdout.
// 	// NOTE: The goroutine will keep reading until the next keystroke before returning.
// 	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
// 	_, _ = io.Copy(os.Stdout, ptmx)

// 	return nil
// }

// Write implements io.Writer.
func (r read_out) Write(p []byte) (n int, err error) {
	if r.handle != nil {
		return r.handle(p)
	}
	gui.Write(p)
	return logFile.Write(p)
}
func setupLogFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

type Pty struct {
	Cmd  *exec.Cmd
	File pty.Pty
	Ch   chan os.Signal
	wch  chan bool
	Rows uint16 // ws_row: Number of rows (in cells).
	Cols uint16 //
}

func (pty *Pty) UpdateSize(Rows uint16, Cols uint16) {
	pty.OsUpdateSize(Rows, Cols)
}

func Ptymain(Args []string) *Pty {
	defer logFile.Close()
	log.SetOutput(logFile)

	// 创建一个新的伪终端
	// lspvi := "/Users/jialaizhu/dev/lspvi/lspvi"
	return RunCommand(Args)
}
func RunNoStdin(Args []string) *Pty {
	c := exec.Command(Args[0])
	c.Args = Args
	f, err := pty.Start(c)
	if err != nil {
		debug.ErrorLog("pty", err)
		return nil
	}
	// var stdout2 read_out
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.
	if err != nil {
		log.Fatal(err)
	}
	ret := &Pty{Cmd: c, File: f, Ch: make(chan os.Signal, 1), wch: make(chan bool, 1)}

	return ret
}

func RunCommand(Args []string) *Pty {
	c := exec.Command(Args[0])
	c.Args = Args
	f, err := pty.Start(c)
	if err != nil {
		log.Panic(err)
	}
	// var stdout2 read_out
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		var stdin2 read_out
		stdin2.handle = func(p []byte) (n int, err error) {
			return f.Write(p)
		}
		io.Copy(stdin2, os.Stdin)
	}()
	ret := &Pty{File: f, Ch: make(chan os.Signal, 1),wch: make(chan bool, 1),}
	ret.Notify()
	go func() {
		for {
			select {
			case <-ret.wch:
				{
					if err := pty.Setsize(ret.File, &pty.Winsize{Rows: ret.Rows, Cols: ret.Cols}); err != nil {
						debug.DebugLogf("pty","error resizing pty: %s", err)
					}
				}
			case <-ret.Ch:
				{
					// if err := pty.InheritSize(os.Stdin, ret.File); err != nil {
					// }
					if err := pty.Setsize(ret.File, &pty.Winsize{Rows: ret.Rows, Cols: ret.Cols}); err != nil {
						log.Printf("error resizing pty: %s", err)
					}
				}
			}
		}
	}()
	return ret
}

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
