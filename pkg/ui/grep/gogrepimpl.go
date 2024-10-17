//go:build !windows

package grep

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"syscall"

	"zen108.com/lspvi/pkg/debug"
)

func PosixRunGrep(grep *Gorep, fpath string, out chan<- GrepInfo) {
	defer func() {
		grep.waitGreps.Done()
	}()
	if grep.IsAbort() {
		return
	}
	// if strings.HasPrefix(fpath, grep.global_prj_root) {
	// 	debug.InfoLog("Ignore ", fpath)
	// }
	grep.filecount++
	file, err := os.Open(fpath)
	if err != nil {
		debug.ErrorLog("grep open error: ", err)
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		debug.ErrorLog("grep state error: ", err)
		return
	}

	mem, err := syscall.Mmap(int(file.Fd()), 0, int(fi.Size()),
		syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {

		return
	}
	defer syscall.Munmap(mem)

	isBinary := verifyBinary(mem)
	if isBinary && !grep.scope.binary {
		return
	}

	buffer := bytes.NewBuffer(mem)
	scanner := bufio.NewScanner(buffer)

	scanner.Split(bufio.ScanLines)
	lineNumber := 0

	var ret = GrepInfo{fpath, lineNumber, "", 0, false}
	for scanner.Scan() {
		lineNumber++
		strline := scanner.Text()
		finded := grep.newFunction1(strline)
		if finded {
			if !grep.just_grep_file {
				if isBinary {
					out <- GrepInfo{fpath, 0, fmt.Sprintf("Binary file %s matches", fpath), 1, false}
					return
				} else {
					out <- GrepInfo{fpath, lineNumber, strline, 1, false}
				}
			} else {
				if ret.Matched == 0 {
					ret.LineNumber = lineNumber
					ret.Line = strline
				}
				ret.Matched++
			}
		}
	}
	if grep.just_grep_file {
		if ret.Matched > 0 {
			out <- ret
		}
	}
}
