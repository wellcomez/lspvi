package grep

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar"
	str "github.com/boyter/go-string"
	"zen108.com/lspvi/pkg/debug"
	gi "zen108.com/lspvi/pkg/ui/gitignore"
)

var GrepTag = "Grep"

type GrepInfo struct {
	Fpath      string
	LineNumber int
	Line       string
	Matched    int
	end        bool
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
	Query       string
	Ignorecase  bool
	Wholeword   bool
	Exclude     bool
	PathPattern string
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
	path_pattern string
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
	opened_file     int
	grep_status     grep_status
	default_gi      gi.Matcher
}

func (g *Gorep) IsRunning() bool {
	return g.grep_status == GrepRunning
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
		grep.CB = nil
		debug.DebugLog(GrepTag, "Run Abort", grep.String())
		return
	}
}

const separator = string(os.PathSeparator)

func (grep *Gorep) Report(chans *channelSet) {
	// var markGrep string
	var waitReports sync.WaitGroup

	chPrint := make(chan GrepOutput)
	chEnd := make(chan bool)

	go func() {
		for {
			select {
			case <-chEnd:
				return
			case msg := <-chPrint:
				if grep.grep_status == GrepRunning {
					grep.CB(grep.id, &msg)
				}
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
		chEnd <- true
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
		if opt.Ignorecase {
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
		// if len(opt.ignore) > 0 {
		// 	if opt.ignorecase {
		// 		opt.ignore = "(?i)" + opt.ignore
		// 	}
		// }
	}
	if len(opt.PathPattern) > 0 {
		base.path_pattern = opt.PathPattern
	}

	base.scope.grep = true
	return base, nil
}

func (grep *Gorep) Kick(fpath string) {
	grep.global_prj_root = fpath
	chsMap := makeChannelSet()
	chsReduce := makeChannelSet()
	grep.waitMaps.Add(1)
	go func() {
		home, _ := os.UserHomeDir()
		ps, _ := gi.ReadIgnoreFile(filepath.Join(home, ".gitignore_global"))
		grep.default_gi = gi.NewMatcher(ps, false)
		grep.mapsend(fpath, chsMap, grep.default_gi)
		grep.waitMaps.Wait()
		closeChannelSet(chsMap)
	}()
	go func() {
		grep.reduce(chsMap, chsReduce)
	}()
	go grep.Report(chsReduce)
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
func (grep *Gorep) String() string {
	status := ""
	switch grep.grep_status {
	case GrepRunning:
		status = "Running"
	case GrepAbort:
		status = "Abort"
	case GrepDone:
		status = "Done"
	}
	return fmt.Sprintln(grep.id, grep.ptnstring, "Opened", grep.opened_file, "status=", status, "Files", grep.filecount, "Line=", grep.count, time.Now().UnixMilli()-grep.begintm)
}
func (grep *Gorep) mapsend(fpath string, chans *channelSet, m gi.Matcher) {
	defer grep.waitMaps.Done()
	debug.TraceLog(GrepTag, "mapsend ", grep.String())
	if grep.IsAbort() {
		return
	}
	/* expand dir */
	list, err := os.ReadDir(fpath)
	if err != nil {
		debug.ErrorLog(GrepTag, "readir error: ", err)
		return
	}
	if data, err := gi.EnterDir(fpath); err == nil && len(data) > 0 {
		m = gi.NewMatcher(grep.default_gi.Patterns(), false)
		m.AddPatterns(data)
	}

	for _, finfo := range list {
		if grep.IsAbort() {
			debug.DebugLog(GrepTag, "mapsend ", grep.String())
			return
		}
		fname := finfo.Name()
		if fname[0] == '.' {
			debug.TraceLog(GrepTag, "ignore:", filepath.Join(fpath, finfo.Name()))
			continue
		}

		path := filepath.Join(fpath, fname)
		is_dir := finfo.IsDir()

		ss := strings.Split(path, separator)
		if m.Match(ss[1:], is_dir) {
			debug.TraceLog(GrepTag, "ignore:", path)
			continue
		}
		if finfo.Type().IsRegular() {
			skip := false
			if len(grep.path_pattern) > 0 {
				skip = true
				if yes, _ := doublestar.Match(grep.path_pattern, path); yes {
					skip = false
				} else if yes, _ := doublestar.Match(grep.path_pattern, finfo.Name()); yes {
					skip = false
				}
			}
			if skip {
				debug.DebugLog(GrepTag, "ignore:"+grep.path_pattern, path)
				continue
			}
		}
		if finfo.IsDir() {
			grep.waitMaps.Add(1)
			go grep.mapsend(path, chans, m)
		} else if finfo.Type().IsRegular() {
			chans.grep <- GrepInfo{path, 0, "", 0, false}
		}
	}
}

func (grep *Gorep) reduce(chsIn *channelSet, chsOut *channelSet) {
	go func(in <-chan GrepInfo, out chan<- GrepInfo) {
		for msg := range in {
			if grep.IsAbort() {
				debug.DebugLog(GrepTag, "reduce abort ", grep.String())
				continue
			}
			grep.waitGreps.Add(1)
			grep.opened_file++
			go grep.grep(msg.Fpath, out)
		}
		grep.waitGreps.Wait()
		close(out)
		debug.DebugLog("reduce done", grep.String())
		if grep.grep_status == GrepRunning {
			grep.grep_status = GrepDone
			if grep.CB != nil {
				grep.CB(grep.id, nil)
			}
		}
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
	grep.opened_file--
}
