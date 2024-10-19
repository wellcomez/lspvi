// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type Differ struct {
	bufer        []string
	changed_line int
}

func (dif *Differ) getChangedLineNumbers(lines2 []string) []int {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(strings.Join(dif.bufer,"\n"), strings.Join(lines2,"\n"), false)
	ret:=[]int{}
	for i, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffDelete {
			ret=append(ret, i+1)
			// fmt.Printf("Deleted line %d: %s\n", i+1, diff.Text)
		} else if diff.Type == diffmatchpatch.DiffInsert {
			ret=append(ret, i+1)
			// fmt.Printf("Inserted line %d: %s\n", i+1, diff.Text)
		}
	}
	return ret
}