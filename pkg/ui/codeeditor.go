package mainui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
	lspcore "zen108.com/lspvi/pkg/lsp"
	fileloader "zen108.com/lspvi/pkg/ui/fileload"
)

type CodeEditor interface {
	//Complete
	HasComplete() bool
	CloseComplete()
	Complete()

	//code change check
	NewChangeChecker() code_change_cheker

	//split
	SplitRight() *CodeView

	//Primitive
	Primitive() tview.Primitive

	//path
	FileName() string
	Path() string

	//handle key event
	handle_key(event *tcell.EventKey) *tcell.EventKey

	//bookmark
	bookmark()

	//lsp
	key_call_in()
	action_goto_declaration()
	action_get_refer()
	action_get_implementation(line *lspcore.OpenOption)
	action_goto_define(line *lspcore.OpenOption)
	Format()
	get_symbol_range(sym lspcore.Symbol) lsp.Range
	LspSymbol() *lspcore.Symbol_file
	TreeSitter() *lspcore.TreeSitter
	get_callin(sym lspcore.Symbol)
	open_picker_refs()

	//selected
	ResetSelection()
	GetSelection() string

	//search
	OnSearch(txt string, whole bool) []SearchPos

	//load
	IsLoading() bool
	Reload()
	Save() error
	LoadBuffer(fileloader.FileLoader)
	LoadFileNoLsp(filename string, line int) error
	LoadFileWithLsp(filename string, line *lsp.Location, focus bool)

	//goto line
	Clear()
	GetLines(begin, end int) []string
	vid() view_id
	goto_line_end()
	goto_line_head()
	goto_location_no_history(loc lsp.Range, update bool, option *lspcore.OpenOption)
	goto_line_history(line int, history bool)
	update_with_line_changed()

	//match
	action_grep_word(selected bool)
	Match()
	OnFindInfile(fzf bool, noloop bool) string
	OnFindInfileWordOption(fzf bool, noloop bool, whole bool) string

	//action
	action_key_up()
	action_key_down()
	action_page_down(bool)
	key_left()
	key_right()
	word_left()
	word_right()

	//undo
	Undo()

	//delete
	deleteline()
	deleteword()
	deltext()

	//copy Paste
	copyline(bool)
	GetCode(lsp.Location) string
	Paste()

	//InsertMode
	InsertMode(yes bool)
	//history
	EditorPosition() *EditorPosition

	//draw
	// DrawNavigationBar(x int, y int, w int, screen tcell.Screen)
}
