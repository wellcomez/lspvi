// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package grep
import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"zen108.com/lspvi/pkg/debug"
)
func RunGrep(grep *Gorep, fpath string, out chan<- GrepInfo) {
	if grep.bAbort {
		return
	}
	defer func() {
		<-semFopenLimit
		grep.waitGreps.Done()
	}()
	if strings.HasPrefix(fpath, grep.global_prj_root) {
		debug.InfoLog("Ignore ", fpath)
	}
	semFopenLimit <- 1
	grep.filecount++
	mem,err:=	os.ReadFile(fpath)
	if err!=nil{
		debug.ErrorLog("grep read error: ", err)
		return
	}
	isBinary := verifyBinary(mem)
	if isBinary && !grep.scope.binary {
		return
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
					return
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
	if grep.just_grep_file {
		if ret.Matched > 0 {
			out <- ret
		}
	}
}



 