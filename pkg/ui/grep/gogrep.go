package grep

import (
	"bytes"
	str "github.com/boyter/go-string"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"zen108.com/lspvi/pkg/debug"
	gi "zen108.com/lspvi/pkg/ui/gitignore"
)

var GrepTag = "Grep"

type GrepInfo struct {
	Fpath      string
	LineNumber int
	Line       string
	Matched    int
}
type contenttype int

const (
	DIR_TYPE contenttype = iota
	FILE_TYPE
	BINARY_TYPE
)

type GrepOutput struct {
	*GrepInfo
	content_type contenttype
}

type channelSet struct {
	grep chan GrepInfo
}

type OptionSet struct {
	G             bool
	Grep_only     bool
	search_binary bool
	ignore        string
	// hidden        bool
	ignorecase bool
	Wholeword  bool
}

type searchScope struct {
	// dir     bool
	// file    bool
	// symlink bool
	grep   bool
	binary bool
}
type grep_status int

const (
	GrepRunning grep_status = iota
	GrepAbort
	GrepDone
)

type Gorep struct {
	pattern         *regexp.Regexp
	ptnstring       string
	useptnstring    bool
	scope           searchScope
	CB              func(taskid int, out *GrepOutput)
	id              int
	waitMaps        sync.WaitGroup
	waitGreps       sync.WaitGroup
	begintm, count  int64
	filecount       int
	just_grep_file  bool
	global_prj_root string
	// report_end      chan bool
	grep_status grep_status
}

func (grep *Gorep) newFunction1(strline string) bool {
	grep.count++
	if grep.useptnstring {
		return len(str.IndexAll(strline, grep.ptnstring, 1)) > 0
	} else {
		return grep.pattern.MatchString(strline)
	}
}
func (grep *Gorep) IsAbort() bool {
	return grep.grep_status == GrepAbort
}
func (grep *Gorep) Abort() {
	switch grep.grep_status {
	case GrepRunning:
		grep.grep_status = GrepAbort
		// grep.report_end <- true
		return
	}
}

const separator = string(os.PathSeparator)

func (grep *Gorep) Report(chans *channelSet, isColor bool) {
	// var markGrep string
	var waitReports sync.WaitGroup

	chPrint := make(chan GrepOutput)
	chPrintEnd := make(chan string)

	go func() {
		for {
			select {
			case msg := <-chPrint:
				grep.CB(grep.id, &msg)
			case <-chPrintEnd:
				if grep.grep_status == GrepRunning {
					grep.CB(grep.id, nil)
				}
				grep.grep_status = GrepDone
			}
		}
	}()

	reporter := func(mark string, chanIf interface{}) {
		defer waitReports.Done()
		switch ch := chanIf.(type) {
		case chan GrepInfo:
			for msg := range ch {
				if grep.IsAbort() {
					continue
				}
				if msg.LineNumber != 0 {
					// decoStr := grep.pattern.ReplaceAllString(msg.line, accent)
					a := GrepOutput{
						// destor: decoStr,
						GrepInfo: &GrepInfo{
							LineNumber: msg.LineNumber,
							Line:       msg.Line,
							Fpath:      msg.Fpath,
							Matched:    msg.Matched,
						},
						content_type: FILE_TYPE,
					}
					chPrint <- a
				} else { // binary file
					a := GrepOutput{
						GrepInfo: &GrepInfo{
							LineNumber: msg.LineNumber,
							Line:       msg.Line,
							Fpath:      msg.Fpath,
						},
						content_type: BINARY_TYPE,
					}
					chPrint <- a
				}
			}
		default:
			break
		}
		grep.Debug("Reporter End")
		chPrintEnd <- mark
	}

	waitReports.Add(1)
	go reporter("grep", chans.grep)
	waitReports.Wait()
}

func NewGorep(id int, pattern string, opt *OptionSet) (*Gorep, error) {
	base := &Gorep{
		pattern:   nil,
		ptnstring: pattern,
		scope: searchScope{
			grep:   false,
			binary: false,
		},
		useptnstring:   true,
		id:             id,
		begintm:        time.Now().UnixMilli(),
		just_grep_file: false,
		// report_end:     make(chan bool),
	}
	base.Debug("NewGrep")

	var err error
	if !base.useptnstring {
		if opt.ignorecase {
			pattern = "(?i)" + pattern
		}
		if opt.Wholeword {
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
		}
	}

	// config search scope
	if opt.G {
		base.scope.grep = true
	}
	if opt.Grep_only {
		// base.scope.file = false
		// base.scope.symlink = false
		base.scope.grep = true
	}
	if opt.search_binary {
		base.scope.binary = true
	}
	return base, nil
}

func (grep *Gorep) Kick(fpath string) *channelSet {
	grep.global_prj_root = fpath
	chsMap := makeChannelSet()
	chsReduce := makeChannelSet()
	grep.waitMaps.Add(1)
	go func() {
		home, _ := os.UserHomeDir()
		ps, _ := gi.ReadIgnoreFile(filepath.Join(home, ".gitignore_global"))
		m := gi.NewMatcher(ps, true)
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
		grep: make(chan GrepInfo),
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
func (grep *Gorep) Debug(s string) {
	debug.InfoLog(GrepTag, s, "Abort=", grep.grep_status, "id=", grep.id, "Files", grep.filecount, "Line=", grep.count, grep.ptnstring, time.Now().UnixMilli()-grep.begintm)
}
func (grep *Gorep) mapsend(fpath string, chans *channelSet, m gi.Matcher) {
	defer grep.waitMaps.Done()
	if grep.IsAbort() {
		debug.DebugLog("Abort Return " + fpath)
		return
	}
	/* expand dir */
	list, err := os.ReadDir(fpath)
	if err != nil {
		debug.ErrorLog(GrepTag, "readir error: ", err)
		return
	}
	m.Enter(fpath)

	for _, finfo := range list {
		if grep.IsAbort() {
			debug.DebugLog("Abort Return " + fpath)
			return
		}
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
			continue
		}
		if finfo.IsDir() {
			grep.waitMaps.Add(1)
			go grep.mapsend(path, chans, m)
		} else if finfo.Type().IsRegular() {
			chans.grep <- GrepInfo{path, 0, "", 0}
		}
	}
}

func (grep *Gorep) reduce(chsIn *channelSet, chsOut *channelSet) {
	go func(in <-chan GrepInfo, out chan<- GrepInfo) {
		for msg := range in {
			grep.waitGreps.Add(1)
			go grep.grep(msg.Fpath, out)
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

func (grep *Gorep) grep(fpath string, out chan<- GrepInfo) {
	//fmt.Fprintf(os.Stderr, "grep mmap error: %v\n", err)
	RunGrep(grep, fpath, out)
}
