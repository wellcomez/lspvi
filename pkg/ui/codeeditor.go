package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
)

type CodeEditor interface {
	HasComplete() bool
	CloseComplete()
	NewChangeChecker() code_change_cheker
	Complete()
	SplitRight() *CodeView
	Clear()
	IsLoading() bool
	GetLines(begin, end int) []string
	GetSelection() string
	OnSearch(txt string, whole bool) []SearchPos
	vid() view_id
	Primitive() tview.Primitive
	Format()
	FileName() string
	Path() string

	update_with_line_changed()

	Save() error

	handle_key(event *tcell.EventKey) *tcell.EventKey

	ResetSelection()

	bookmark()

	key_call_in()
	action_goto_declaration()
	action_get_refer()
	action_get_implementation(line *lspcore.OpenOption)
	action_goto_define(line *lspcore.OpenOption)

	// openfile(filename string, onload func()) error
	Reload()
	LoadBuffer(data []byte, filename string)
	LoadFileNoLsp(filename string, line int) error

	LoadFileWithLsp(filename string, line *lsp.Location, focus bool)

	goto_location_no_history(loc lsp.Range, update bool, option *lspcore.OpenOption)
	goto_line_history(line int, history bool)

	get_symbol_range(sym lspcore.Symbol) lsp.Range
	LspSymbol() *lspcore.Symbol_file

	TreeSitter() *lspcore.TreeSitter
	Match()

	action_key_up()
	action_key_down()
	action_page_down(bool)
	key_left()
	key_right()
	word_left()
	word_right()

	Undo()
	deleteline()
	deleteword()
	copyline(bool)
	Paste()
	deltext()

	goto_line_end()
	goto_line_head()
	get_callin(sym lspcore.Symbol)

	open_picker_refs()
	InsertMode(yes bool)

	action_grep_word(selected bool)
	OnFindInfile(fzf bool, noloop bool) string
	OnFindInfileWordOption(fzf bool, noloop bool, whole bool) string

	EditorPosition() *EditorPosition

	DrawNavigationBar(x int, y int, w int, screen tcell.Screen)
}
