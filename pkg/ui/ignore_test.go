// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"path/filepath"
	"testing"
	"time"
)

func Test_mutli(t *testing.T) {
	root := "/chrome/buildcef/chromium/src/"
	var yes=false
	git := NewGitIgnore("/chrome/buildcef/chromium/src/")
	yes = git.Ignore(filepath.Join(root, "PRESUBMIT_test_mocks.py"))
	if yes {
		t.Error("should not ignore")
	}
	// git.Ignore(filepath.Join(root, ".git/config"))
	b:=time.Now()
	yes = git.Ignore(filepath.Join(root, "compile_commands.json"))
	t.Logf("time:%d", time.Since(b))
	if !yes {
		t.Error("should ignore")
	}

}
