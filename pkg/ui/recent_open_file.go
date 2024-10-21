// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui
type recent_open_file struct {
	list *customlist
	*view_link
	filelist []string
	main     *mainui
	Name     string
}

func (r *recent_open_file) add(filename string) {
	go func() {
		GlobalApp.QueueUpdateDraw(func() {
			for _, v := range r.filelist {
				if v == filename {
					return
				}
			}
			filepath := filename
			r.filelist = append(r.filelist, filename)
			filename = trim_project_filename(filename, global_prj_root)
			r.list.AddItem(filename, "", func() {
				r.main.OpenFileHistory(filepath, nil)
			})
		})
	}()
}
func new_recent_openfile(m *mainui) *recent_open_file {
	return &recent_open_file{
		list:      new_customlist(false),
		view_link: &view_link{id: view_recent_open_file},
		filelist:  []string{},
		main:      m,
		Name:      "Opened Files",
	}
}