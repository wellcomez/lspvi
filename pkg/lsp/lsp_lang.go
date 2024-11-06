package lspcore

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tectiv3/go-lsp"
)

type lsplang interface {
	Launch_Lsp_Server(core *lspcore, wk WorkSpace) error
	InitializeLsp(core *lspcore, wk WorkSpace) error
	IsSource(filename string) bool
	Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool
	IsMe(filename string) bool
	CompleteHelpCallback(lsp.CompletionList, *Complete, error)
	LspHelp(*lspcore) (LspSignatureHelp, LspCompleteUtil, error)
}
type lsp_lang_common struct {
}
type Document struct {
	Value string `json:"value"`
}

func (v *Document) Parser(a []byte) error {
	if err := json.Unmarshal(a, v); err != nil {
		return err
	}
	if len(v.Value) == 0 {
		return errors.New("no value")
	}
	return nil
}
func (a lsp_lang_common) LspHelp(*lspcore) (h LspSignatureHelp, c LspCompleteUtil, err error) {
	err = fmt.Errorf("not support")
	return
}
func (a lsp_lang_common) CompleteHelpCallback(cl lsp.CompletionList, ret *Complete, err error) {
	document := []string{}
	for index := range cl.Items {

		v := cl.Items[index]
		var text = []string{
			v.Label,
			v.Detail}
		var doc Document
		if doc.Parser(v.Documentation) == nil {
			text = append(text, "//"+doc.Value)
		}
		document = append(document, strings.Join(text, "\n"))
	}
	ret.Result = &CompleteResult{Document: document}
}
