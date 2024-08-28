package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
)

func maintty() {
	// 创建一个主从PTY对
	master, slave, err := ptyOpen()
	if err != nil {
		fmt.Println("Failed to open PTY:", err)
		return
	}
	defer master.Close()
	defer slave.Close()

	// 创建子进程
	cmd := exec.Command("/bin/ls", "-l")

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
		io.Copy(os.Stdout, master)
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
	if _, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), ioctlTtyAlloc, 0); e != 0 {
		err = fmt.Errorf("ioctl(TTY_ALLOC): %v", e)
		return
	}

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

	return
}

const (
	// ioctlTtyAlloc 是用于分配一个新的PTY的ioctl命令
	ioctlTtyAlloc = 0x5412 // TIOCGPT
)
