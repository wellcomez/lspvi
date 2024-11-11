package ptyproxy

// https://github.com/iyzyi/aiopty
// https://github.com/photostorm/pty
// https://github.com/creack/pty/pull/155
import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

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

type LspPty interface {
	File() io.ReadWriteCloser
	UpdateSize(Rows uint16, Cols uint16)
	Kill() error
	Notify()
	Pid() string
}
type PtyCmd struct {
	cmd              *exec.Cmd
	file             io.ReadWriteCloser
	ws_change_signal chan os.Signal
	set_size_changed chan bool
	rows             uint16 // ws_row: Number of rows (in cells).
	cols             uint16 //
}

func (pty *PtyCmd) File() io.ReadWriteCloser {
	return pty.file
}

func (pty *PtyCmd) UpdateSize(Rows uint16, Cols uint16) {
	if Rows != pty.rows || Cols != pty.cols {
		pty.OsUpdateSize(Rows, Cols)
	}
}

func Ptymain(Args []string) LspPty {
	defer logFile.Close()
	log.SetOutput(logFile)

	// 创建一个新的伪终端
	// lspvi := "/Users/jialaizhu/dev/lspvi/lspvi"
	return RunCommand(Args)
}
func RunNoStdin(Args []string) LspPty {
	if Use_aio_pty {
		cmdline := strings.Join(Args, " ")
		return NewAioptyPtyCmd(cmdline, false)
	} else {
		return RunNoStdinCreak(Args)
	}
}
func RunNoStdinCreak(Args []string) *PtyCmd {
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
	ret := &PtyCmd{cmd: c, file: f, ws_change_signal: make(chan os.Signal, 1), set_size_changed: make(chan bool, 1)}
	ret.monitorSizeChanged(f)
	return ret
}
func (t *PtyCmd) Kill() error {
	if t.cmd != nil {
		return t.cmd.Process.Kill()
	}
	return fmt.Errorf("not cmd")
}
func (pty *PtyCmd) Pid() (pid string) {
	if pty.cmd != nil {
		pid = fmt.Sprintf("%d",
			pty.cmd.Process.Pid)
	}
	return pid
}

var Use_aio_pty = true

func RunCommand(Args []string) LspPty {
	// var stdout2 read_out
	// Best effort.
	// if err := pty.InheritSize(os.Stdin, ret.File); err != nil {
	// }
	if Use_aio_pty {
		return NewAioptyPtyCmd(strings.Join(Args, " "), true)
	}
	return use_creak_pty(Args)
}

func use_creak_pty(Args []string) LspPty {
	c := exec.Command(Args[0])
	c.Args = Args
	f, err := pty.Start(c)
	if err != nil {
		log.Panic(err)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()
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
	ret := &PtyCmd{file: f, ws_change_signal: make(chan os.Signal, 1), set_size_changed: make(chan bool, 1)}
	ret.Notify()

	ret.monitorSizeChanged(f)
	return ret
}

func (ptycmd *PtyCmd) monitorSizeChanged(f pty.Pty) {
	go func() {
		for {
			select {
			case <-ptycmd.set_size_changed:
				{
					if err := pty.Setsize(f, &pty.Winsize{Rows: ptycmd.rows, Cols: ptycmd.cols}); err != nil {
						debug.DebugLogf("pty", "error resizing pty: %s", err)
					}
				}
			case <-ptycmd.ws_change_signal:
				{

					if err := pty.Setsize(f, &pty.Winsize{Rows: ptycmd.rows, Cols: ptycmd.cols}); err != nil {
						log.Printf("error resizing pty: %s", err)
					}
				}
			}
		}
	}()
}
