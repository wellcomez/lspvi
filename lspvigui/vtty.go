//go:build ignore
// +build ignore

package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync/atomic"

	// "strings"
	"syscall"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

type read struct {
}

const consoleDevice string = "/dev/tty"

var devPrefixes = [...]string{"/dev/pts/", "/dev/"}

var tty atomic.Value

func ttyname() string {
	if cached := tty.Load(); cached != nil {
		return cached.(string)
	}

	var stderr syscall.Stat_t
	if syscall.Fstat(2, &stderr) != nil {
		return ""
	}

	for _, prefix := range devPrefixes {
		files, err := os.ReadDir(prefix)
		if err != nil {
			continue
		}

		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				continue
			}
			if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat.Rdev == stderr.Rdev {
				value := prefix + file.Name()
				tty.Store(value)
				return value
			}
		}
	}
	return ""
}
func (r *LightRenderer) fd() int {
	return int(r.ttyin.Fd())
}
func (r *LightRenderer) initPlatform() (err error) {
	r.origState, err = term.MakeRaw(r.fd())
	return err
}

// TtyIn returns terminal device to read user input
func TtyIn() (*os.File, error) {
	return openTtyIn()
}

// TtyIn returns terminal device to write to
func TtyOut() (*os.File, error) {
	return openTtyOut()
}
func openTty(mode int) (*os.File, error) {
	in, err := os.OpenFile(consoleDevice, mode, 0)
	if err != nil {
		tty := ttyname()
		if len(tty) > 0 {
			if in, err := os.OpenFile(tty, mode, 0); err == nil {
				return in, nil
			}
		}
		return nil, errors.New("failed to open " + consoleDevice)
	}
	return in, nil
}

type LightRenderer struct {
	ttyin     *os.File
	ttyout    *os.File
	origState *term.State
}
type TermSize struct {
	Lines    int
	Columns  int
	PxWidth  int
	PxHeight int
}

func (r *LightRenderer) Size() TermSize {
	ws, err := unix.IoctlGetWinsize(int(r.ttyin.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return TermSize{}
	}
	return TermSize{int(ws.Row), int(ws.Col), int(ws.Xpixel), int(ws.Ypixel)}
}
func openTtyIn() (*os.File, error) {
	return openTty(syscall.O_RDONLY)
}

func openTtyOut() (*os.File, error) {
	return openTty(syscall.O_WRONLY)
}

// Write implements io.Writer.
func (r read) Write(p []byte) (n int, err error) {
	fmt.Println("read: ", len(p), string(p))
	return len(p), nil
}

func maintty(cmd *exec.Cmd) {

	// 创建一个主从PTY对
	master, slave, err := ptyOpen()
	if err != nil {
		fmt.Println("Failed to open PTY:", err)
		return
	}
	defer master.Close()
	defer slave.Close()

	// 创建子进程
	// cmd := exec.Command("/bin/ls", "-l")

	// 设置子进程的文件描述符
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // 创建一个新的会话
	}

	// 将从设备作为子进程的输入输出
	cmd.Stdin = slave
	cmd.Stdout = slave
	cmd.Stderr = slave

	// 启动子进程
	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start process:", err)
		return
	}

	// 从主设备读取数据
	go func() {
		var r read
		io.Copy(r, master)
	}()

	// 向主设备写入数据
	io.Copy(master, os.Stdin)

	// 等待子进程退出
	if err := cmd.Wait(); err != nil {
		fmt.Println("Process exited with error:", err)
		return
	}

	fmt.Println("Subprocess exited successfully")
}

// ptyOpen 创建一个PTY对
func ptyOpen() (master *os.File, slave *os.File, err error) {
	var fds [2]int
	// if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), ioctlTtyAlloc, 0); e != 0 {
	// 	err = fmt.Errorf("ioctl(TTY_ALLOC): %v", e)
	// 	return
	// }
	ttyin, err := openTtyIn()
	if err != nil {
		fmt.Println("ttyin: ", err)
	}

	ttyout, err := openTtyOut()
	if err != nil {
		fmt.Println("ttyout", err)
	}
	h := LightRenderer{
		ttyin:  ttyin,
		ttyout: ttyout,
	}
	err = h.initPlatform()
	if err != nil {
		fmt.Println("initPlatform", err)
	}
	fmt.Println(h.Size())

	if err := syscall.Pipe2(fds[:], syscall.O_NONBLOCK); err != nil {
		log.Fatal(err)
		return nil, nil, err
	}
	if fds[0] == -1 || fds[1] == -1 {
		err = fmt.Errorf("openpty failed")
		return
	}

	master = os.NewFile(uintptr(fds[0]), "pty-master")
	slave = os.NewFile(uintptr(fds[1]), "pty-slave")
	// s,err:=term.MakeRaw(int(slave.Fd()))
	// if err!=nil{
	// 	fmt.Println("MakeRaw",err)
	// }
	// fmt.Println(s)
	return
}
