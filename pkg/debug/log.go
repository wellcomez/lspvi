package debug

import (
	"fmt"
	"log"
)

type log_level int

const (
	log_level_error log_level = iota
	log_level_warn
	log_level_info
	log_level_debug
	log_level_trace
)

var loglevel log_level = log_level_debug

func commonprintln(debug, tag string, args ...interface{}) {
	s := fmt.Sprint(args...)
	log.Println(fmt.Sprintf("[%s][%s]", debug, tag), s)
}
func DebugLog(tag string, v ...any) {
	if loglevel >= log_level_debug {
		commonprintln(tag, "DEBUG", v...)
	}
}
func TraceLogf(tag, format string, v ...any) {
	if loglevel >= log_level_trace {
		CommonLogf(tag, "TRACE", format, v...)
	}
}

func TraceLog(tag string, v ...any) {
	if loglevel >= log_level_trace {
		commonprintln(tag, "TRACE", v...)
	}
}
func InfoLogf(tag, format string, v ...any) {
	if loglevel >= log_level_info {
		CommonLogf(tag, "INFO ", format, v...)
	}
}

func InfoLog(tag string, v ...any) {
	if loglevel >= log_level_info {
		commonprintln(tag, "INFO ", v...)
	}
}
func ErrorLog(tag string, v ...any) {
	if loglevel >= log_level_error {
		commonprintln(tag, "ERROR", v...)
	}
}
func WarnLog(tag string, v ...interface{}) {
	if loglevel >= log_level_warn {
		commonprintln(tag, "WARN ", v...)
	}
}
func CommonLogf(debug, tag, format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	log.Println(debug, tag, s)
}
func DebugLogf(tag, format string, v ...any) {
	if loglevel >= log_level_debug {
		CommonLogf(tag, "DEBUG", format, v...)
	}
}
func ErrorLogf(tag, format string, v ...any) {
	if loglevel >= log_level_error {
		CommonLogf(tag, "ERROR", format, v...)
	}
}
func WarnLogf(tag, format string, v ...any) {
	if loglevel >= log_level_warn {
		CommonLogf(tag, "WARN ", format, v...)
	}
}
