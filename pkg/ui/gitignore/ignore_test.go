// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package gitignore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"zen108.com/lspvi/pkg/debug"
	// "github.com/wellsjo/SuperSearch/src/logger"
)

var testDir string

var ignoreFile1 = ".gitignore"

func TestMain(m *testing.M) {
	testDir, err := ioutil.TempDir("", "ss-test")
	if err != nil {
		debug.ErrorLog(err.Error())
	}
	ioutil.WriteFile(filepath.Join(testDir, "test"), []byte{}, 0644)
	defer func() {
		os.Remove(testDir)
	}()
	os.Exit(m.Run())
}

func TestGitignore(t *testing.T) {
	// patterns := gitignore.ReadIgnoreFile(testDir)
}
