package mainui

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"syscall"
)
func RunGrep(grep *gorep, fpath string, out chan<- grepInfo) bool {
	return RunGrepLinux(grep, fpath, out)
}
func RunGrepLinux(grep *gorep, fpath string, out chan<- grepInfo) bool {
	defer func() {
		<-semFopenLimit
		grep.waitGreps.Done()
	}()

	if yes, err := isSubdir(lspviroot.root, fpath); err != nil && yes {
		return true
	}
	semFopenLimit <- 1
	file, err := os.Open(fpath)
	if err != nil {
		log.Printf("grep open error: %v\n", err)
		return true
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Printf("grep stat error: %v\n", err)
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
		if grep.pattern.MatchString(strline) {
			if isBinary {
				out <- grepInfo{fpath, 0, fmt.Sprintf("Binary file %s matches", fpath)}
				return true
			} else {
				if grep.ignorePattern != nil && grep.ignorePattern.MatchString(strline) {
					continue
				}
				out <- grepInfo{fpath, lineNumber, strline}
			}
		}
	}
	return false
}
