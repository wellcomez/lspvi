package lspcore

import (
	"fmt"
	"os/exec"
)

type lsp_cpp struct {
	lsp_base
}

func (l lsp_cpp) InitializeLsp(wk workroot) error {
	result, err := l.core.Initialize(wk)
	if err != nil {
		return err
	}
	if result.ServerInfo.Name == "clangd" {
		return nil
	}
	return fmt.Errorf("%s", result.ServerInfo.Name)
}

// Launch_Lsp_Server implements lspclient.
func (l lsp_cpp) Launch_Lsp_Server() error {
	root := "--compile-commands-dir=" + l.wk.path
	l.core.cmd = exec.Command("clangd", root)
	return l.core.Lauch_Lsp_Server(l.core.cmd)
}
