package mainui

import fileloader "zen108.com/lspvi/pkg/ui/fileload"

// "github.com/sergi/go-diff"
func (code *CodeView) DiffLine(beign int) bool {
	if ret, err := code.Diff(beign, beign); err == nil {
		return len(ret) > 0
	}
	return false
}
func (code *CodeView) Diff(beign, end int) (diff []int, err error) {

	if load, e := fileloader.Loader.GetFile(code.Path(), false); e == nil {
		old_len := len(load.Lines.Lines)
		newline := load.Buff.Lines(beign, end+1)
		for i, v := range newline {
			lineno := i + beign
			if lineno < old_len {
				if load.Lines.Lines[i] != v {
					diff = append(diff, lineno)
				}
			} else {
				diff = append(diff, lineno)
			}
		}
	} else {
		err = e
	}
	return
}
