package lspcore

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tectiv3/go-lsp"
)

type LspUtil struct {
	Signature LspSignatureHelp
	Complete  LspCompleteUtil
}
type lsplang interface {
	Launch_Lsp_Server(core *lspcore, wk WorkSpace) error
	InitializeLsp(core *lspcore, wk WorkSpace) error
	IsSource(filename string) bool
	Resolve(sym lsp.SymbolInformation, symfile *Symbol_file) bool
	IsMe(filename string) bool
	CompleteHelpCallback(lsp.CompletionList, *Complete, error)
	LspHelp(*lspcore) (LspUtil, error)
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
func (a lsp_lang_common) LspHelp(core *lspcore) (ret LspUtil, err error) {
	var h LspSignatureHelp
	var c LspCompleteUtil
	err = fmt.Errorf("not support")
	if core.CapabilitiesStatus.CompletionProvider != nil {
		c.TriggerChar = core.CapabilitiesStatus.CompletionProvider.TriggerCharacters
	}
	if core.CapabilitiesStatus.SignatureHelpProvider != nil {
		h.TriggerChar = core.CapabilitiesStatus.SignatureHelpProvider.TriggerCharacters
	}
	ret = LspUtil{
		Signature: h,
		Complete:  c,
	}
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
