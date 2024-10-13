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

func PosixRunGrep(grep *gorep, fpath string, out chan<- grepInfo) bool {
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

	for scanner.Scan() {
		lineNumber++
		strline := scanner.Text()
		x := grep.newFunction1(strline)
		if x {
			if isBinary {
				out <- grepInfo{fpath, 0, fmt.Sprintf("Binary file %s matches", fpath)}
				return true
			} else {
				// if grep.ignorePattern != nil && grep.ignorePattern.MatchString(strline) {
				// 	continue
				// }
				out <- grepInfo{fpath, lineNumber, strline}
			}
			if grep.just_grep_file{
				break
			}
		}
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
