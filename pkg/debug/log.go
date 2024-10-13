package debug

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
)

type log_level int

const (
	log_level_error log_level = iota
	log_level_warn
	log_level_info
	log_level_debug
	log_level_trace
)
const TagUI = "UI"

var loglevel log_level = log_level_debug

func log_prefix(tag, debug string) string {
	_, file, line, _ := runtime.Caller(2)
	x := fmt.Sprintf("[%-5s][%s] %s:%d", debug, tag, filepath.Base(file), line)
	return x
}
func DebugLog(tag string, v ...any) {
	if loglevel >= log_level_debug {
		log.Println(log_prefix(tag, "DEBUG"), fmt.Sprint(v...))
	}
}
func TraceLogf(tag, format string, v ...any) {
	if loglevel >= log_level_trace {
		log.Printf(log_prefix(tag, "TRACE")+format, v...)
	}
}

func TraceLog(tag string, v ...any) {
	if loglevel >= log_level_trace {
		log.Println(log_prefix(tag, "TRACE"), fmt.Sprint(v...))
	}
}
func InfoLogf(tag, format string, v ...any) {
	if loglevel >= log_level_info {
		log.Printf(log_prefix(tag, "INFO ")+format, v...)
	}
}

func InfoLog(tag string, v ...any) {
	if loglevel >= log_level_info {
		log.Println(log_prefix(tag, "INFO"), fmt.Sprint(v...))
	}
}
func ErrorLog(tag string, v ...any) {
	if loglevel >= log_level_error {
		log.Println(log_prefix(tag, "ERROR"), fmt.Sprint(v...))
	}
}
func WarnLog(tag string, v ...interface{}) {
	if loglevel >= log_level_warn {
		log.Println(log_prefix(tag, "WARN"), fmt.Sprint(v...))
	}
}
func DebugLogf(tag, format string, v ...any) {
	if loglevel >= log_level_debug {
		log.Printf(log_prefix(tag, "DEBUG")+format, v...)
	}
}
func ErrorLogf(tag, format string, v ...any) {
	if loglevel >= log_level_error {
		log.Printf(log_prefix(tag, "ERROR")+format, v...)
	}
}
func WarnLogf(tag, format string, v ...any) {
	if loglevel >= log_level_warn {
		log.Printf(log_prefix(tag, "WARN ")+format, v...)
	}
}
