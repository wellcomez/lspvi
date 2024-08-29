package mainui

import (
	"encoding/json"
	"os"
	"path/filepath"
)
type uisize struct {
	Width int
	Hide  bool
}
type editor_area_layout struct {
	Code uisize
	File uisize
	Sym  uisize
	main *mainui
	name string
}

func (e editor_area_layout) save() error {
	buf, err := json.Marshal(e)
	if err == nil {
		return os.WriteFile(e.name, buf, 0666)
	}
	return err
}
func (e *editor_area_layout) increate(id view_id, a int) {
	link := id.to_view_link(e.main)
	if link == nil {
		return
	}
	s := e.get_view_uisize(id)
	if s == nil {
		return
	}
	link.width += a
	s.Width = link.width
	e.config_to_ui()
}

func (e *editor_area_layout) get_view_uisize(id view_id) *uisize {
	var s *uisize
	switch id {
	case view_code:
		s = &e.Code
	case view_file:
		s = &e.File
	case view_outline_list:
		s = &e.Sym
	}
	return s
}
func (e *editor_area_layout) Hide(id view_id, yes bool) {
	s := e.get_view_uisize(id)
	if s == nil {
		return
	}
	s.Hide = yes
	e.config_to_ui()
}
func new_editor_area_config(m *mainui, root *workdir) *editor_area_layout {
	e := &editor_area_layout{main: m}
	e.Code = uisize{Width: 20, Hide: false}
	e.File = uisize{Width: 5, Hide: false}
	e.Sym = uisize{Width: 5, Hide: false}
	file := filepath.Join(root.export, "editlayout.json")
	e.name = file
	buf, err := os.ReadFile(file)
	var v editor_area_layout
	if err != nil {
		return e
	}
	err = json.Unmarshal(buf, &v)
	if err == nil {
		e.Code = v.Code
		e.File = v.File
		e.Sym = v.Sym
	}
	e.config_to_ui()
	return e
}

func (e *editor_area_layout) config_to_ui() {
	m := e.main
	m.codeview.width = e.Code.Width
	m.codeview.hide = e.Code.Hide

	m.fileexplorer.width = e.File.Width
	m.fileexplorer.hide = e.File.Hide

	m.symboltree.width = e.Sym.Width
	m.symboltree.hide = e.Sym.Hide
	e.save()
}
func (e *editor_area_layout) zoom(zoomin bool, viewid view_id) {
	m := e.main
	add := 1
	if zoomin {
		add = -1
	}
	switch viewid {
	case view_outline_list, view_file:
		{
			m._editor_area_layout.increate(viewid, add)
		}
	default:
		return
	}
	e.update_editerea_layout()

}

func (e *editor_area_layout) toggle_view(id view_id) {
	m := e.main
	switch id {
	case view_file, view_outline_list:
		if link := id.to_view_link(m); link != nil {
			m._editor_area_layout.Hide(id, !link.hide)
		} else {
			return
		}
	default:
		return
	}
	e.update_editerea_layout()
}

func (e editor_area_layout) update_editerea_layout() {
	m := e.main
	m.layout.editor_area.Clear()
	if !m.fileexplorer.hide {
		m.layout.editor_area.AddItem(m.fileexplorer.view, 0, m.fileexplorer.width, false)
	}
	m.layout.editor_area.AddItem(m.codeview.view, 0, m.codeview.width, false)
	if !m.symboltree.hide {
		m.layout.editor_area.AddItem(m.symboltree.view, 0, m.symboltree.width, false)
	}
}