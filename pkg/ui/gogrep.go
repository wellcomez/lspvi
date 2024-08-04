package mainui

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	// "code.google.com/p/go.crypto/ssh/terminal"
)

const version = "0.2.5"

type grepInfo struct {
	fpath      string
	lineNumber int
	line       string
}
type contenttype int

const (
	DIR_TYPE contenttype = iota
	FILE_TYPE
	SYMLINK_TYPE
	BINARY_TYPE
	STRING_TYPE
)

type grep_output struct {
	*grepInfo
	destor       string
	content_type contenttype
}

type channelSet struct {
	dir     chan string
	file    chan string
	symlink chan string
	grep    chan grepInfo
}

type optionSet struct {
	v             bool
	g             bool
	grep_only     bool
	search_binary bool
	ignore        string
	hidden        bool
	ignorecase    bool
}

type searchScope struct {
	dir     bool
	file    bool
	symlink bool
	grep    bool
	binary  bool
	hidden  bool
}

type gorep struct {
	pattern       *regexp.Regexp
	ignorePattern *regexp.Regexp
	scope         searchScope
	cb            func(taskid int, out *grep_output)
	id            int
}

var semFopenLimit chan int

const maxNumOfFileOpen = 10

var waitMaps sync.WaitGroup
var waitGreps sync.WaitGroup

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

func verifyColor() bool {
	// fd := os.Stdout.Fd()
	// isTerm := terminal.IsTerminal(int(fd))
	// return isTerm
	return false
}

func test_main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var opt optionSet
	flag.BoolVar(&opt.v, "V", false, "show version.")
	flag.BoolVar(&opt.g, "g", false, "enable grep.")
	flag.BoolVar(&opt.search_binary, "search-binary", false, "search binary files for matches on grep enable.")
	flag.BoolVar(&opt.grep_only, "grep-only", false, "enable grep and disable file search.")
	flag.StringVar(&opt.ignore, "ignore", "", "pattern is ignored.")
	flag.BoolVar(&opt.hidden, "hidden", false, "search hidden files.")
	flag.BoolVar(&opt.ignorecase, "ignorecase", false, "ignore case distinctions in pattern.")
	flag.Parse()

	if opt.v {
		fmt.Printf("version: %s\n", version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		// usage()
	}
	pattern := flag.Arg(0)
	fpath := "."
	if flag.NArg() >= 2 {
		fpath = strings.TrimRight(flag.Arg(1), separator)
	}

	g, err := newGorep(1, pattern, &opt)
	if err == nil {
		chans := g.kick(fpath)
		g.report(chans, verifyColor())
	}
}

const (
	DIR_COLOR     = "\x1b[36m"
	FILE_COLOR    = "\x1b[34m"
	SYMLINK_COLOR = "\x1b[35m"
	GREP_COLOR    = "\x1b[32m"
	HIT_COLOR     = "\x1b[32m"
	NORM_COLOR    = "\x1b[39m"
	BOLD_DECO     = "\x1b[1m"
	NORM_DECO     = "\x1b[0m"
)

func (grep *gorep) end(o grep_output) {
	if grep.cb != nil {
		grep.cb(grep.id, &o)
	}
}
func (grep *gorep) report(chans *channelSet, isColor bool) {
	var markMatch string
	var markDir string
	var markFile string
	var markSymlink string
	var markGrep string
	if isColor {
		markMatch = BOLD_DECO + HIT_COLOR + "$0" + NORM_COLOR + NORM_DECO
		markDir = DIR_COLOR + "[Dir ]" + NORM_COLOR
		markFile = FILE_COLOR + "[File]" + NORM_COLOR
		markSymlink = SYMLINK_COLOR + "[SymL]" + NORM_COLOR
		markGrep = GREP_COLOR + "[Grep]" + NORM_COLOR
	} else {
		markMatch = "$0"
		markDir = "[Dir ]"
		markFile = "[File]"
		markSymlink = "[SymL]"
		markGrep = "[Grep]"
	}

	var waitReports sync.WaitGroup

	chPrint := make(chan grep_output)
	// printer
	go func() {
		for {
			msg := <-chPrint
			grep.end(msg)
			// os.Stdout.Write(<-chPrint)
		}
	}()

	reporter := func(mark string, accent string, chanIf interface{}) {
		defer waitReports.Done()
		switch ch := chanIf.(type) {
		case chan string:
			for msg := range ch {
				decoStr := grep.pattern.ReplaceAllString(msg, accent)
				a := grep_output{
					destor:       decoStr,
					content_type: STRING_TYPE,
				}
				chPrint <- a
			}
		case chan grepInfo:
			for msg := range ch {
				if msg.lineNumber != 0 {
					decoStr := grep.pattern.ReplaceAllString(msg.line, accent)
					a := grep_output{
						destor:       decoStr,
						grepInfo:     &msg,
						content_type: FILE_TYPE,
					}
					chPrint <- a
				} else { // binary file
					a := grep_output{
						grepInfo:     &msg,
						content_type: BINARY_TYPE,
					}
					chPrint <- a
				}
			}
		default:
			log.Printf("reporter type error!\n")
			return
		}
	}

	waitReports.Add(4)
	go reporter(markDir, markMatch, chans.dir)
	go reporter(markFile, markMatch, chans.file)
	go reporter(markSymlink, markMatch, chans.symlink)
	go reporter(markGrep, markMatch, chans.grep)
	waitReports.Wait()
}

func newGorep(id int, pattern string, opt *optionSet) (*gorep, error) {
	base := &gorep{
		pattern:       nil,
		ignorePattern: nil,
		scope: searchScope{
			dir:     true,
			file:    true,
			symlink: true,
			grep:    false,
			binary:  false,
			hidden:  false,
		},
		id: id,
	}

	// config regexp
	if opt.ignorecase {
		pattern = "(?i)" + pattern
	}

	var err error
	base.pattern, err = regexp.Compile(pattern)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "%v\n", err)
		// os.Exit(-1)
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
		base.scope.dir = false
		base.scope.file = false
		base.scope.symlink = false
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
		waitMaps.Add(1)
		grep.mapsend(fpath, chsMap)
		waitMaps.Wait()
		closeChannelSet(chsMap)
	}()

	go func() {
		grep.reduce(chsMap, chsReduce)
	}()
	return chsReduce
}

func makeChannelSet() *channelSet {
	return &channelSet{
		dir:     make(chan string),
		file:    make(chan string),
		symlink: make(chan string),
		grep:    make(chan grepInfo),
	}
}

func closeChannelSet(chans *channelSet) {
	close(chans.dir)
	close(chans.file)
	close(chans.symlink)
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
	defer waitMaps.Done()

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
			if grep.scope.dir {
				chans.dir <- fullpath
			}
			waitMaps.Add(1)
			go grep.mapsend(fullpath, chans)
		case mode&os.ModeSymlink != 0:
			if grep.scope.symlink {
				chans.symlink <- fpath + separator + fname
			}
		case mode&ignoreFlag == 0:
			fullpath := fpath + separator + fname
			if grep.scope.file {
				chans.file <- fullpath
			}
			if grep.scope.grep {
				chans.grep <- grepInfo{fullpath, 0, ""}
			}
		default:
			continue
		}
	}
}

func (grep *gorep) reduce(chsIn *channelSet, chsOut *channelSet) {
	filter := func(in <-chan string, out chan<- string) {
		for msg := range in {
			if grep.pattern.MatchString(path.Base(msg)) {
				out <- msg
			}
		}
		close(out)
	}

	// directory
	go filter(chsIn.dir, chsOut.dir)

	// file
	go filter(chsIn.file, chsOut.file)

	// symlink
	go filter(chsIn.symlink, chsOut.symlink)

	// grep
	go func(in <-chan grepInfo, out chan<- grepInfo) {
		for msg := range in {
			waitGreps.Add(1)
			go grep.grep(msg.fpath, out)
		}
		waitGreps.Wait()
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

func (grep *gorep) grep(fpath string, out chan<- grepInfo) {
	defer func() {
		<-semFopenLimit
		waitGreps.Done()
	}()

	semFopenLimit <- 1
	file, err := os.Open(fpath)
	if err != nil {
		log.Printf("grep open error: %v\n", err)
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Printf("grep stat error: %v\n", err)
		return
	}

	mem, err := syscall.Mmap(int(file.Fd()), 0, int(fi.Size()),
		syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "grep mmap error: %v\n", err)
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

	for scanner.Scan() {
		lineNumber++
		strline := scanner.Text()
		if grep.pattern.MatchString(strline) {
			if isBinary {
				out <- grepInfo{fpath, 0, fmt.Sprintf("Binary file %s matches", fpath)}
				return
			} else {
				if grep.ignorePattern != nil && grep.ignorePattern.MatchString(strline) {
					continue
				}
				out <- grepInfo{fpath, lineNumber, strline}
			}
		}
	}
}
