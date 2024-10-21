package fileloader

import (
	"os"

	"github.com/pgavlin/femto"
)

// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3
var Loader FileLoaderMgr

type FileLoader struct {
	Buff     *femto.Buffer
	FileName string
}
type FileLoaderMgr struct {
	filemap map[string]FileLoader
}

func NewDataFileLoad(data []byte, file string) FileLoader {
	return FileLoader{Buff: femto.NewBufferFromString(string(data), file), FileName: file}
}
func (fm *FileLoaderMgr) GetFile(filename string, reload bool) (ret FileLoader, err error) {
	if fm.filemap == nil {
		fm.filemap = map[string]FileLoader{}
	}
	if !reload {
		if a, yes := fm.filemap[filename]; yes {
			return a, err
		}
	}
	ret, err = fm.LoadFile(filename)
	if err == nil {
		fm.filemap[filename] = ret
	}
	return
}
func (fm *FileLoaderMgr) LoadFile(filename string) (ret FileLoader, err error) {
	if data, err := os.ReadFile(filename); err == nil {
		return FileLoader{
			FileName: filename,
			Buff:     femto.NewBufferFromString(string(data), filename),
		}, nil
	}
	return FileLoader{FileName: filename}, err
}
