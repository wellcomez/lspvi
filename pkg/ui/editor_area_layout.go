package mainui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type resizable_layout interface {
	zoom(zoomin bool, viewid view_id)
}
type editor_area_layout struct {
	main *mainui
	name string
}

func (e editor_area_layout) save() error {
	buf, err := json.Marshal(editlayout_config_data{
		Sym:  *e.main.symboltree.view_link,
		File: *e.main.fileexplorer.view_link,
		Code: *e.main.codeview.view_link,
	})
	if err == nil {
		return os.WriteFile(e.name, buf, 0666)
	}
	return err
}
func (e *editor_area_layout) increate(id view_id, a int) {
	link := id.to_view_link(e.main)
	if link != nil {
		link.Width += a
		link.Width=max(1,link.Width)
		e.save()
	}
}

func (e *editor_area_layout) Hide(id view_id, yes bool) {
	link := id.to_view_link(e.main)
	if link != nil {
		link.Hide = yes
		e.save()
	}
}

type editlayout_config_data struct {
	Code, File, Sym view_link
}

func new_editor_area_config(m *mainui, root *workdir) *editor_area_layout {
	e := &editor_area_layout{main: m}
	m.codeview.Width = 20
	m.codeview.Hide = false

	m.fileexplorer.Width = 5
	m.fileexplorer.Hide = false

	m.symboltree.Width = 5
	m.symboltree.Hide = false

	file := filepath.Join(root.export, "editlayout.json")
	e.name = file
	buf, err := os.ReadFile(file)
	var v editlayout_config_data
	if err != nil {
		return e
	}
	err = json.Unmarshal(buf, &v)
	if err == nil {
		m.codeview.Width = v.Code.Width
		m.codeview.Hide = v.Code.Hide

		m.fileexplorer.Width = v.File.Width
		m.fileexplorer.Hide = v.File.Hide

		m.symboltree.Width = v.Sym.Width
		m.symboltree.Hide = v.Sym.Hide
	}
	return e
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
			m._editor_area_layout.Hide(id, !link.Hide)
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
	if !m.fileexplorer.Hide {
		m.layout.editor_area.AddItem(m.fileexplorer.view, 0, m.fileexplorer.Width, false)
	}
	m.layout.editor_area.AddItem(m.codeview.view, 0, m.codeview.Width, false)
	if !m.symboltree.Hide {
		m.layout.editor_area.AddItem(m.symboltree.view, 0, m.symboltree.Width, false)
	}
}
