package mainui

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"zen108.com/lspvi/pkg/debug"
	gi "zen108.com/lspvi/pkg/ui/gitignore"
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
	pattern        *regexp.Regexp
	ignorePattern  *regexp.Regexp
	ptnstring      string
	useptnstring   bool
	scope          searchScope
	cb             func(taskid int, out *grep_output)
	id             int
	bAbort         bool
	waitMaps       sync.WaitGroup
	waitGreps      sync.WaitGroup
	begintm, count int64
	filecount      int
	just_grep_file bool
}

func (grep *gorep) abort() {
	grep.bAbort = true
}

var semFopenLimit chan int

const maxNumOfFileOpen = 10

const separator = string(os.PathSeparator)

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
			<-chPrintEnd
			grep.Debug("grep end")
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
				if grep.bAbort {
					grep.Debug("grep abort in report loop")
					break
				}
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
		ptnstring:     pattern,
		scope: searchScope{
			grep:   false,
			binary: false,
			hidden: false,
		},
		useptnstring:   true,
		id:             id,
		begintm:        time.Now().UnixMilli(),
		just_grep_file: true,
	}
	base.Debug("NewGrep")

	// config regexp
	if opt.ignorecase {
		pattern = "(?i)" + pattern
	}

	var err error
	if !base.useptnstring {
		if opt.wholeword {
			pattern = `\b` + pattern + `\b`
		}
		base.pattern, err = regexp.Compile(pattern)
		if err != nil {
			debug.ErrorLog(GrepTag, "regexp error", err)
			return nil, err
		}
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
		home, _ := os.UserHomeDir()
		ps, _ := gi.ReadIgnoreFile(filepath.Join(home, ".gitignore_global"))
		m := gi.NewMatcher(ps)
		grep.mapsend(fpath, chsMap, m)
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

//	func verifyHidden(fpath string) bool {
//		byteStr := []byte(path.Base(fpath))
//		// don't consider current directory(./) and parent directory(../)
//		if '.' == byteStr[0] {
//			return true
//		}
//		return false
//	}

// isHidden checks if a file or directory is hidden.
func (grep *gorep) Debug(s string) {
	debug.InfoLog(GrepTag, s, "Abort=", grep.bAbort, "id=", grep.id, "Files", grep.filecount, "Line=", grep.count, grep.ptnstring, time.Now().UnixMilli()-grep.begintm)
}
func (grep *gorep) mapsend(fpath string, chans *channelSet, m gi.Matcher) {
	defer grep.waitMaps.Done()
	if grep.bAbort {
		grep.Debug("Abort")

		return
	}
	/* expand dir */
	list, err := os.ReadDir(fpath)
	if err != nil {
		debug.ErrorLog(GrepTag, "readir error: ", err)
		return
	}
	ignore_path := filepath.Join(fpath, ".gitignore")
	ps, _ := gi.ReadIgnoreFile(ignore_path)
	if len(ps) > 0 {
		m = gi.NewMatcher(ps)
		debug.InfoLog(GrepTag, "new gitignore:", ignore_path)
	}

	for _, finfo := range list {
		fname := finfo.Name()
		if fname[0] == '.' {
			debug.InfoLog(GrepTag, "ignore:", filepath.Join(fpath, finfo.Name()))
			continue
		}

		path := filepath.Join(fpath, fname)
		is_dir := finfo.IsDir()

		ss := strings.Split(path, separator)
		if m.Match(ss[1:], is_dir) {
			debug.InfoLog(GrepTag, "ignore:", path)
		}
		if finfo.IsDir() {
			grep.waitMaps.Add(1)
			go grep.mapsend(path, chans, m)
		} else if finfo.Type().IsRegular() {
			chans.grep <- grepInfo{path, 0, ""}
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

func (grep *gorep) grep(fpath string, out chan<- grepInfo) {
	//fmt.Fprintf(os.Stderr, "grep mmap error: %v\n", err)
	RunGrep(grep, fpath, out)
}
