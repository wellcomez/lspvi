package mainui

import (
	"log"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tectiv3/go-lsp"
)

type CodeContextMenu struct {
	code *CodeView
}

// on_mouse implements context_menu_handle.
func (menu CodeContextMenu) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	code := menu.code
	root := code.view

	code.view.get_click_line_inview(event)
	right_menu_data := code.right_menu_data
	pos := root.event_to_cursor_position(event) // pos = avoid_position_overflow(root, pos)

	if action == tview.MouseRightClick {

		//select line
		selected := code.get_selected_lines()
		right_menu_data.previous_selection = selected

		Loc := code.view.tab_loc(pos)
		right_menu_data.local_changed = root.Cursor.Loc != (Loc)
		//save cursor loc
		cursor_data := *root.Cursor
		right_menu_data.rightmenu_loc = Loc

		//get selected text
		word_select_cursor := code.SelectWordFromCopyCursor(cursor_data)
		_, s := get_codeview_text_loc(root.View, word_select_cursor.CurSelection[0], word_select_cursor.CurSelection[1])
		menu.code.right_menu_data.select_text = s
		menu.code.right_menu_data.selection_range = text_loc_to_range(word_select_cursor.CurSelection)

		// move cursor to mouse postion
		code.set_loc(Loc)
		update_selection_menu(code)

	}
	return action, event
}

// getbox implements context_menu_handle.
func (code CodeContextMenu) getbox() *tview.Box {
	if code.code.main == nil {
		return nil
	}
	main := code.code.main
	if code.code.id == view_code_below {
		if main.Tab().activate_tab_id != view_code_below {
			return nil
		}
	}
	return code.code.view.Box
}

// menuitem implements context_menu_handle.
func (code CodeContextMenu) menuitem() []context_menu_item {
	return code.code.rightmenu_items
}
func update_selection_menu(code *CodeView) {
	main := code.main
	toggle_file_view := "Toggle file view"
	fileexplorer := main.FileExplore()
	if !fileexplorer.Hide {
		toggle_file_view = "Hide file view"
	}
	toggle_outline := "Toggle outline view"
	x := main.OutLineView()
	if !x.Hide {
		toggle_outline = "Hide outline view"
	}
	tty := main.Mode().tty
	menudata := code.right_menu_data
	items := []context_menu_item{
		{item: create_menu_item("Reference"), handle: func() {

			if menudata.SelectInEditor(code.view.Cursor) {
				main.get_refer(menudata.selection_range, code.Path())
				main.ActiveTab(view_quickview, false)
			}
		}},
		{item: create_menu_item("Implementation"), handle: func() {

			if menudata.SelectInEditor(code.view.Cursor) {
				main.get_implementation(menudata.selection_range, code.Path(), nil)
				main.ActiveTab(view_quickview, false)
			}
		}},
		{item: create_menu_item("Goto define"), handle: func() {
			menudata.SelectInEditor(code.view.Cursor)
			main.get_define(menudata.selection_range, code.Path(), nil)
			main.ActiveTab(view_quickview, false)
		}},
		{item: create_menu_item("Call incoming"), handle: func() {
			if menudata.SelectInEditor(code.view.Cursor) {
				loc := lsp.Location{
					URI:   lsp.NewDocumentURI(code.Path()),
					Range: menudata.selection_range,
				}
				main.get_callin_stack_by_cursor(loc, code.Path())
				main.ActiveTab(view_callin, false)
			}
		}},
		{item: create_menu_item("Open in explorer"), handle: func() {
			// ret.filename
			dir := filepath.Dir(code.Path())
			fileexplorer.ChangeDir(dir)
			fileexplorer.FocusFile(code.Path())
		}},
		{item: create_menu_item("-------------"), handle: func() {
		}},
		{item: create_menu_item("Bookmark"), handle: func() {
			code.bookmark()
		}, hide: menudata.previous_selection.emtry()},
		{item: create_menu_item("Save Selection"), handle: func() {
			code.save_selection(menudata.previous_selection.selected_text)
		}},
		{item: create_menu_item("Search Selection"), handle: func() {
			sss := menudata.previous_selection
			main.OnSearch(search_option{sss.selected_text, true, true, false})
			main.ActiveTab(view_quickview, false)
		}, hide: menudata.previous_selection.emtry()},
		{item: create_menu_item("Search"), handle: func() {
			sss := menudata.select_text
			menudata.SelectInEditor(code.view.Cursor)
			main.OnSearch(search_option{sss, true, true, false})
			main.ActiveTab(view_quickview, false)
		}, hide: len(menudata.select_text) == 0},
		{item: create_menu_item("Grep word"), handle: func() {
			rightmenu_select_text := menudata.select_text
			main.qf_grep_word(rightmenu_select_text)
			menudata.SelectInEditor(code.view.Cursor)
		}, hide: len(menudata.select_text) == 0},
		{item: create_menu_item("Copy Selection"), handle: func() {
			selected := menudata.previous_selection
			data := selected.selected_text
			if selected.emtry() {
				data = menudata.select_text
			}
			code.main.CopyToClipboard(data)

		}, hide: menudata.previous_selection.emtry()},
		{item: create_menu_item("-"), handle: func() {
		}},
		{item: create_menu_item(toggle_file_view), handle: func() {
			main.toggle_view(view_file)
		}, hide: code.id != view_code},
		{item: create_menu_item(toggle_outline), handle: func() {
			main.toggle_view(view_outline_list)
		}},

		{item: create_menu_item("-"), handle: func() {
		}},
		SplitDown(code),
		SplitRight(code),
		SplitClose(code),
		{item: create_menu_item("-"), handle: func() {
		}},
		{
			item: create_menu_item("External open "),
			handle: func() {
				filename := code.Path()
				yes, err := isDirectory(filename)
				if err != nil {
					return
				}
				log.Println("external open tty=", tty)
				if proxy != nil {
					proxy.open_in_web(filename)
				} else {
					if !yes {
						openfile(filename)
					}
				}
			},
		},
		{item: create_menu_item("-"), handle: func() {
		}, hide: !tty},
		{item: create_menu_item("Zoom-in Browser"), handle: func() {
			main.ZoomWeb(false)
		}, hide: !tty},
		{item: create_menu_item("Zoom-out Browser"), handle: func() {
			main.ZoomWeb(true)
		}, hide: !tty},
	}
	code.rightmenu_items = addjust_menu_width(items)
}

func SplitDown(code *CodeView) context_menu_item {
	main := code.main
	return context_menu_item{item: create_menu_item("SplitDown"), handle: func() {
		if code.id >= view_code {
			main.Codeview2().LoadFileWithLsp(code.Path(), nil, false)
		}
	}}
}
