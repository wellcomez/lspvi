package mainui

import (
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
)
type CodeContextMenu struct {
	code *CodeView
}

// on_mouse implements context_menu_handle.
func (menu CodeContextMenu) on_mouse(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	code := menu.code
	root := code.view

	code.view.get_click_line_inview(event)
	posX, posY := event.Position()
	right_menu_data := code.right_menu_data
	yOffset := code.view.yOfffset()
	xOffset := code.view.xOffset()
	// offsetx:=3
	pos := mouse_event_pos{
		Y: posY + root.Topline - yOffset,
		X: posX - int(xOffset),
	}
	// pos = avoid_position_overflow(root, pos)

	if action == tview.MouseRightClick {
		selected := code.get_selected_lines()
		right_menu_data.previous_selection = selected
		// code.rightmenu.text = root.Cursor.GetSelection()
		cursor := *root.Cursor
		Loc := code.view.tab_loc(pos)
		code.set_loc(Loc)
		cursor.SetSelectionStart(femto.Loc{X: pos.X, Y: pos.Y})
		right_menu_data.rightmenu_loc = cursor.CurSelection[0]
		// log.Println("before", cursor.CurSelection)
		loc := code.SelectWord(cursor)
		_, s := get_codeview_text_loc(root.View, loc.CurSelection[0], loc.CurSelection[1])
		menu.code.right_menu_data.select_text = s
		menu.code.right_menu_data.selection_range = text_loc_to_range(loc.CurSelection)
		// code.get_selected_lines()
		// code.rightmenu_select_text = root.Cursor.GetSelection()
		// code.rightmenu_select_range = code.convert_curloc_range(code.view.Cursor.CurSelection)
		// log.Println("after ", code.view.Cursor.CurSelection)
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
		if main.tab.activate_tab_id != view_code_below {
			return nil
		}
	}
	return code.code.view.Box
}

// menuitem implements context_menu_handle.
func (code CodeContextMenu) menuitem() []context_menu_item {
	return code.code.rightmenu_items
}
