package mainui

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"zen108.com/lspvi/pkg/debug"
	// "code.google.com/p/go.crypto/ssh/terminal"
)

var GrepTag = "Grep"

type grepInfo struct {
	fpath      string
	lineNumber int
	line       string
}
type contenttype int

const (
	DIR_TYPE contenttype = iota
	FILE_TYPE
	BINARY_TYPE
)

type grep_output struct {
	*grepInfo
	// destor       string
	content_type contenttype
}

type channelSet struct {
	// dir     chan string
	// file    chan string
	// symlink chan string
	grep chan grepInfo
}

type optionSet struct {
	v             bool
	g             bool
	grep_only     bool
	search_binary bool
	ignore        string
	hidden        bool
	ignorecase    bool
	wholeword     bool
}

type searchScope struct {
	// dir     bool
	// file    bool
	// symlink bool
	grep   bool
	binary bool
	hidden bool
}

type gorep struct {
	pattern       *regexp.Regexp
	ignorePattern *regexp.Regexp
	scope         searchScope
	cb            func(taskid int, out *grep_output)
	id            int
	bAbort        bool
	waitMaps      sync.WaitGroup
	waitGreps     sync.WaitGroup
}

func (grep *gorep) abort() {
	grep.bAbort = true
}

var semFopenLimit chan int

const maxNumOfFileOpen = 10

const separator = string(os.PathSeparator)

// func usage() {
// 	fmt.Fprintf(os.Stderr, `gorep is find and grep tool.

// Version: %s

// Usage:

//     gorep [options] pattern [path]

// The options are:

//     -V              : print gorep version
//     -g              : enable grep
//     -grep-only      : enable grep and disable file search
//     -search-binary  : search binary files for matches on grep enable
//     -ignore pattern : pattern is ignored
//     -hidden         : search hidden files
//     -ignorecase     : ignore case distinctions in pattern
// `, version)
// 	os.Exit(-1)
// }

func init() {
	semFopenLimit = make(chan int, maxNumOfFileOpen)
}

func (grep *gorep) end(o grep_output) {
	if grep.cb != nil {
		grep.cb(grep.id, &o)
	}
}
func (grep *gorep) report(chans *channelSet, isColor bool) {
	// var markGrep string
	var waitReports sync.WaitGroup

	chPrint := make(chan grep_output)
	chPrintEnd := make(chan string)
	go func() {
		// var i = 0
		for {
			msg := <-chPrintEnd
			debug.InfoLog(GrepTag, "chPrintEnd--------->", msg)
			if grep.cb != nil {
				grep.cb(grep.id, nil)
			}
		}
	}()
	// printer
	go func() {
		for {
			msg := <-chPrint
			grep.end(msg)
		}
	}()

	reporter := func(mark string, chanIf interface{}) {
		defer waitReports.Done()
		switch ch := chanIf.(type) {
		case chan grepInfo:
			for msg := range ch {
				if msg.lineNumber != 0 {
					// decoStr := grep.pattern.ReplaceAllString(msg.line, accent)
					a := grep_output{
						// destor: decoStr,
						grepInfo: &grepInfo{
							lineNumber: msg.lineNumber,
							line:       msg.line,
							fpath:      msg.fpath,
						},
						content_type: FILE_TYPE,
					}
					chPrint <- a
				} else { // binary file
					a := grep_output{
						grepInfo: &grepInfo{
							lineNumber: msg.lineNumber,
							line:       msg.line,
							fpath:      msg.fpath,
						},
						content_type: BINARY_TYPE,
					}
					chPrint <- a
				}
			}
		default:
			debug.ErrorLog("reporter type error!")
			return
		}
		chPrintEnd <- mark
	}

	waitReports.Add(1)
	go reporter("grep", chans.grep)
	waitReports.Wait()
}

func newGorep(id int, pattern string, opt *optionSet) (*gorep, error) {
	base := &gorep{
		pattern:       nil,
		ignorePattern: nil,
		scope: searchScope{
			grep:   false,
			binary: false,
			hidden: false,
		},
		id: id,
	}

	// config regexp
	if opt.ignorecase {
		pattern = "(?i)" + pattern
	}

	var err error
	if opt.wholeword {
		pattern = `\b` + pattern + `\b`
	}
	base.pattern, err = regexp.Compile(pattern)
	if err != nil {
		debug.ErrorLog(GrepTag, "regexp error", err)
		return nil, err
	}

	if len(opt.ignore) > 0 {
		if opt.ignorecase {
			opt.ignore = "(?i)" + opt.ignore
		}
		base.ignorePattern, err = regexp.Compile(opt.ignore)
		if err != nil {
			return nil, err
		}
	}

	// config search scope
	if opt.g {
		base.scope.grep = true
	}
	if opt.grep_only {
		// base.scope.file = false
		// base.scope.symlink = false
		base.scope.grep = true
	}
	if opt.search_binary {
		base.scope.binary = true
	}
	if opt.hidden {
		base.scope.hidden = true
	}

	return base, nil
}

func (grep *gorep) kick(fpath string) *channelSet {
	chsMap := makeChannelSet()
	chsReduce := makeChannelSet()

	go func() {
		grep.waitMaps.Add(1)
		grep.mapsend(fpath, chsMap)
		grep.waitMaps.Wait()
		closeChannelSet(chsMap)
	}()

	go func() {
		grep.reduce(chsMap, chsReduce)
	}()
	return chsReduce
}

func makeChannelSet() *channelSet {
	return &channelSet{
		grep: make(chan grepInfo),
	}
}

func closeChannelSet(chans *channelSet) {
	// close(chans.dir)
	// close(chans.file)
	// close(chans.symlink)
	close(chans.grep)
}

func verifyHidden(fpath string) bool {
	byteStr := []byte(path.Base(fpath))
	// don't consider current directory(./) and parent directory(../)
	if '.' == byteStr[0] {
		return true
	}
	return false
}

func (grep *gorep) mapsend(fpath string, chans *channelSet) {
	defer grep.waitMaps.Done()
	if grep.bAbort {
		return
	}
	/* expand dir */
	list, err := ioutil.ReadDir(fpath)
	if err != nil {
		log.Printf("dive error: %v\n", err)
		return
	}

	const ignoreFlag = os.ModeDir | os.ModeAppend | os.ModeExclusive | os.ModeTemporary |
		os.ModeSymlink | os.ModeDevice | os.ModeNamedPipe | os.ModeSocket |
		os.ModeSetuid | os.ModeSetgid | os.ModeCharDevice | os.ModeSticky

	for _, finfo := range list {
		mode := finfo.Mode()
		fname := finfo.Name()
		if !grep.scope.hidden && verifyHidden(fname) {
			continue
		}
		if grep.ignorePattern != nil && grep.ignorePattern.MatchString(fname) {
			continue
		}
		switch true {
		case mode&os.ModeDir != 0:
			fullpath := fpath + separator + fname
			// if grep.scope.dir {
			// 	chans.dir <- fullpath
			// }
			grep.waitMaps.Add(1)
			go grep.mapsend(fullpath, chans)
		case mode&os.ModeSymlink != 0:
			// if grep.scope.symlink {
			// 	chans.symlink <- fpath + separator + fname
			// }
			continue
		case mode&ignoreFlag == 0:
			fullpath := fpath + separator + fname
			if grep.scope.grep {
				chans.grep <- grepInfo{fullpath, 0, ""}
			}
		default:
			continue
		}
	}
}

func (grep *gorep) reduce(chsIn *channelSet, chsOut *channelSet) {
	go func(in <-chan grepInfo, out chan<- grepInfo) {
		for msg := range in {
			grep.waitGreps.Add(1)
			go grep.grep(msg.fpath, out)
		}
		grep.waitGreps.Wait()
		close(out)
	}(chsIn.grep, chsOut.grep)
}

// Charactor code 0x00 - 0x08 is control code (ASCII)
func verifyBinary(buf []byte) bool {
	var b []byte
	if len(buf) > 256 {
		b = buf[:256]
	} else {
		b = buf
	}
	if bytes.IndexFunc(b, func(r rune) bool { return r < 0x09 }) != -1 {
		return true
	}
	return false
}
func isSubdir(parentPath, childPath string) (bool, error) {
	cleanParent := filepath.Clean(parentPath)
	cleanChild := filepath.Clean(childPath)

	relPath, err := filepath.Rel(cleanParent, cleanChild)
	if err != nil {
		return false, err
	}

	// 如果相对路径以父路径开始，则表示 child 是 parent 的子目录
	return !filepath.IsAbs(relPath) && !strings.HasPrefix(relPath, ".."), nil
}
func (grep *gorep) grep(fpath string, out chan<- grepInfo) {
	//fmt.Fprintf(os.Stderr, "grep mmap error: %v\n", err)
	RunGrep(grep, fpath, out)
}
