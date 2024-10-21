// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package grep

func RunGrep(grep *Gorep, fpath string, out chan<- GrepInfo) {
	PosixRunGrep(grep, fpath, out)
}
