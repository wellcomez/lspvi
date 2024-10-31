package mainui

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"zen108.com/lspvi/pkg/debug"
	lspcore "zen108.com/lspvi/pkg/lsp"
)



func (qf *quickfix_history) save_history(
	root string,
	data qf_history_data, add bool,
) error {
	dir := qf.qfdir
	uid := ""
	if add {
		uid = search_key_uid(data.Key)
		uid = strings.ReplaceAll(uid, root, "")
		hexbuf := md5.Sum([]byte(uid))
		uid = hex.EncodeToString(hexbuf[:])
		data.UID = uid
	} else {
		uid = data.UID
	}
	filename := filepath.Join(dir, uid+".json")
	if !add {
		if last, err := qf_history_ReadLast(qf.Wk); err == nil {
			if last.UID == data.UID {
				os.RemoveAll(qf.last)
			}
		}
		if len(data.UID) != 0 {
			return os.Remove(filename)
		} else if data.Type == data_callin {
			return os.RemoveAll(data.Key.File)
		} else {
			return fmt.Errorf("uid is empty")
		}
	}
	buf, error := json.Marshal(data)
	if error != nil {
		return error
	}
	os.WriteFile(qf.last, buf, 0666)
	return os.WriteFile(filename, buf, 0666)
}

type quickfix_history struct {
	Wk    lspcore.WorkSpace
	last  string
	qfdir string
}


func qf_history_ReadLast(wk lspcore.WorkSpace) (*qf_history_data, error) {
	h, err := new_qf_history(wk)
	if err != nil {
		return nil, err
	}
	buf, err := os.ReadFile(h.last)
	if err != nil {
		return nil, err
	}
	var ret qf_history_data
	err = json.Unmarshal(buf, &ret)
	if err == nil {
		return &ret, nil

	}
	return nil, err
}

func new_qf_history(Wk lspcore.WorkSpace) (*quickfix_history, error) {
	qf := &quickfix_history{
		Wk:   Wk,
		last: filepath.Join(Wk.Export, "quickfix_last.json"),
	}
	qfdir, err := qf.InitDir()
	qf.qfdir = qfdir
	if err != nil {
		debug.ErrorLog("save ", err)
		return nil, err
	}
	return qf, nil
}
func (h *quickfix_history) Load() ([]qf_history_data, error) {
	var ret = []qf_history_data{}
	dir, err := h.InitDir()
	if err != nil {
		return ret, err
	}
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return ret, err
	}
	for _, v := range dirs {
		if v.IsDir() {
			continue
		}
		filename := filepath.Join(dir, v.Name())
		buf, err := os.ReadFile(filename)
		if err != nil {
			continue
		}
		var result qf_history_data
		err = json.Unmarshal(buf, &result)
		if err != nil {
			continue
		}
		ret = append(ret, result)
	}
	umlDir := filepath.Join(h.Wk.Export, "uml")
	dirs, err = os.ReadDir(umlDir)
	if err != nil {
		return ret, err
	}
	for _, dir := range dirs {
		var result = qf_history_data{
			Type: data_callin,
			Key: SearchKey{&lspcore.SymolSearchKey{
				Key:  dir.Name(),
				File: filepath.Join(umlDir, dir.Name()),
			}, nil},
		}
		ret = append(ret, result)

	}
	return ret, nil
}


func (qf *quickfix_history) InitDir() (string, error) {
	Dir := filepath.Join(qf.Wk.Export, "qf")
	if checkDirExists(Dir) {
		return Dir, nil
	}
	err := os.Mkdir(Dir, 0755)
	return Dir, err
}