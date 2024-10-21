// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type logview struct {
	*view_link
	log    *tview.TextView
	lineno int
}

func new_log_view(main *mainui) *logview {
	ret := &logview{
		view_link: &view_link{id: view_log, up: view_code, right: view_callin},
		log:       tview.NewTextView(),
	}
	ret.log.SetDynamicColors(true)
	return ret
}
func (log *logview) clean() {
	log.log.SetText("")
}
func (log *logview) update_log_view(s string) {
	log.lineno++
	t := log.log.GetText(true)
	customLayout := "2006-01-02 15:04:05.000"
	h:=fmt.Sprintf("%d: %v", log.lineno, time.Now().Format(customLayout))
	
	h=fmt.Sprintf("[#8080ff]%s[#ffffff]",h)
	x := t + h+"\n"
	log.log.SetText(
		x + s + "\n")
}
