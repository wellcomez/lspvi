package fileloader

import (
	"os"

	"github.com/pgavlin/femto"
)

// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

type FileLoader struct {
	Buff     *femto.Buffer
	FileName string
}
type FileLoaderMgr struct {
	filemap map[string]FileLoader
}

func (fm *FileLoaderMgr) GetFile(filename string) FileLoader {
	if f, yes := fm.filemap[filename]; yes {
		return f
	} else {
		f := fm.LoadFile(filename)
		fm.filemap[filename] = f
		return f
	}
}
func (fm *FileLoaderMgr) LoadFile(filename string) FileLoader {
	if data, err := os.ReadFile(filename); err == nil {
		return FileLoader{
			FileName: filename,
			Buff:     femto.NewBufferFromString(string(data), filename),
		}
	}
	return FileLoader{FileName: filename}
}
