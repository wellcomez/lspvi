package fileloader

import (
	"os"

	"github.com/pgavlin/femto"
)

// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3
var Loader FileLoaderMgr

type Lines struct {
	Lines []string
}
type FileLoader struct {
	Buff     *femto.Buffer
	FileName string
	Lines    *Lines
}
type FileLoaderMgr struct {
	filemap map[string]FileLoader
}

func NewDataFileLoad(data []byte, file string) FileLoader {
	ret := FileLoader{Buff: femto.NewBufferFromString(string(data), file), FileName: file}
	ret.Lines = &Lines{

		ret.Buff.Lines(0, ret.Buff.LinesNum()-1),
	}
	return ret
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
	if data, e := os.ReadFile(filename); e == nil {
		ret = FileLoader{
			FileName: filename,
			Buff:     femto.NewBufferFromString(string(data), filename),
		}
		ret.Lines = &Lines{
			ret.Buff.Lines(0, ret.Buff.LinesNum()-1)}
		return
	}
	return FileLoader{FileName: filename}, err
}
