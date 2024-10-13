//go:build !windows

package mainui

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"

	str "github.com/boyter/go-string"
	"zen108.com/lspvi/pkg/debug"
)

func PosixRunGrep(grep *gorep, fpath string, out chan<- GrepInfo) bool {
	if grep.bAbort {
		return true
	}
	defer func() {
		<-semFopenLimit
		grep.waitGreps.Done()
	}()
	if strings.HasPrefix(fpath, global_prj_root) {
		debug.InfoLog("Ignore ", fpath)
	}
	semFopenLimit <- 1
	grep.filecount++
	file, err := os.Open(fpath)
	if err != nil {
		debug.ErrorLog("grep open error: ", err)
		return true
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		debug.ErrorLog("grep state error: ", err)
		return true
	}

	mem, err := syscall.Mmap(int(file.Fd()), 0, int(fi.Size()),
		syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {

		return true
	}
	defer syscall.Munmap(mem)

	isBinary := verifyBinary(mem)
	if isBinary && !grep.scope.binary {
		return true
	}

	buffer := bytes.NewBuffer(mem)
	scanner := bufio.NewScanner(buffer)

	scanner.Split(bufio.ScanLines)
	lineNumber := 0

	var ret = GrepInfo{fpath, lineNumber, "", 0}
	for scanner.Scan() {
		lineNumber++
		strline := scanner.Text()
		finded := grep.newFunction1(strline)
		if finded {
			if !grep.just_grep_file {
				if isBinary {
					out <- GrepInfo{fpath, 0, fmt.Sprintf("Binary file %s matches", fpath), 1}
					return true
				} else {
					out <- GrepInfo{fpath, lineNumber, strline, 1}
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
	if ret.Matched > 0 {
		out <- ret
	}
	return false
}

func (grep *gorep) newFunction1(strline string) bool {
	grep.count++
	if grep.useptnstring {
		return len(str.IndexAll(strline, grep.ptnstring, 1)) > 0
	} else {
		return grep.pattern.MatchString(strline)
	}
}
